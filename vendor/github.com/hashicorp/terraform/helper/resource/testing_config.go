package resource

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/config/hcl2shim"
	"github.com/hashicorp/terraform/states"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/tfdiags"
)

// testStepConfig runs a config-mode test step
func testStepConfig(
	opts terraform.ContextOpts,
	state *terraform.State,
	step TestStep) (*terraform.State, error) {
	return testStep(opts, state, step)
}

func testStep(
	opts terraform.ContextOpts,
	state *terraform.State,
	step TestStep) (*terraform.State, error) {
	// Pre-taint any resources that have been defined in Taint, as long as this
	// is not a destroy step.
	if !step.Destroy {
		if err := testStepTaint(state, step); err != nil {
			return state, err
		}
	}

	mod, err := testModule(opts, step)
	if err != nil {
		return state, err
	}

	var stepDiags tfdiags.Diagnostics

	// Build the context
	opts.Config = cfg
	opts.State, err = terraform.ShimLegacyState(state)
	if err != nil {
		return nil, err
	}

	opts.Destroy = step.Destroy
	ctx, stepDiags := terraform.NewContext(&opts)
	if stepDiags.HasErrors() {
		return state, fmt.Errorf("Error initializing context: %s", stepDiags.Err())
	}
	if stepDiags := ctx.Validate(); len(stepDiags) > 0 {
		if stepDiags.HasErrors() {
			return state, errwrap.Wrapf("config is invalid: {{err}}", stepDiags.Err())
		}

		log.Printf("[WARN] Config warnings:\n%s", stepDiags)
	}

	// Refresh!
	newState, stepDiags := ctx.Refresh()
	// shim the state first so the test can check the state on errors

	state, err = shimNewState(newState, step.providers)
	if err != nil {
		return nil, err
	}
	if stepDiags.HasErrors() {
		return state, newOperationError("refresh", stepDiags)
	}

	// If this step is a PlanOnly step, skip over this first Plan and subsequent
	// Apply, and use the follow up Plan that checks for perpetual diffs
	if !step.PlanOnly {
		// Plan!
		if p, stepDiags := ctx.Plan(); stepDiags.HasErrors() {
			return state, newOperationError("plan", stepDiags)
		} else {
			log.Printf("[WARN] Test: Step plan: %s", legacyPlanComparisonString(newState, p.Changes))
		}

		// We need to keep a copy of the state prior to destroying
		// such that destroy steps can verify their behavior in the check
		// function
		stateBeforeApplication := state.DeepCopy()

		// Apply the diff, creating real resources.
		newState, stepDiags = ctx.Apply()
		// shim the state first so the test can check the state on errors
		state, err = shimNewState(newState, step.providers)
		if err != nil {
			return nil, err
		}
		if stepDiags.HasErrors() {
			return state, newOperationError("apply", stepDiags)
		}

		// Run any configured checks
		if step.Check != nil {
			if step.Destroy {
				if err := step.Check(stateBeforeApplication); err != nil {
					return state, fmt.Errorf("Check failed: %s", err)
				}
			} else {
				if err := step.Check(state); err != nil {
					return state, fmt.Errorf("Check failed: %s", err)
				}
			}
		}
	}

	// Now, verify that Plan is now empty and we don't have a perpetual diff issue
	// We do this with TWO plans. One without a refresh.
	var p *plans.Plan
	if p, stepDiags = ctx.Plan(); stepDiags.HasErrors() {
		return state, newOperationError("follow-up plan", stepDiags)
	}
	if !p.Changes.Empty() {
		if step.ExpectNonEmptyPlan {
			log.Printf("[INFO] Got non-empty plan, as expected:\n\n%s", legacyPlanComparisonString(newState, p.Changes))
		} else {
			return state, fmt.Errorf(
				"After applying this step, the plan was not empty:\n\n%s", legacyPlanComparisonString(newState, p.Changes))
		}
	}

	// And another after a Refresh.
	if !step.Destroy || (step.Destroy && !step.PreventPostDestroyRefresh) {
		newState, stepDiags = ctx.Refresh()
		if stepDiags.HasErrors() {
			return state, newOperationError("follow-up refresh", stepDiags)
		}

		state, err = shimNewState(newState, step.providers)
		if err != nil {
			return nil, err
		}
	}
	if p, stepDiags = ctx.Plan(); stepDiags.HasErrors() {
		return state, newOperationError("second follow-up refresh", stepDiags)
	}
	empty := p.Changes.Empty()

	// Data resources are tricky because they legitimately get instantiated
	// during refresh so that they will be already populated during the
	// plan walk. Because of this, if we have any data resources in the
	// config we'll end up wanting to destroy them again here. This is
	// acceptable and expected, and we'll treat it as "empty" for the
	// sake of this testing.
	if step.Destroy && !empty {
		empty = true
		for _, change := range p.Changes.Resources {
			if change.Addr.Resource.Resource.Mode != addrs.DataResourceMode {
				empty = false
				break
			}
		}
	}

	if !empty {
		if step.ExpectNonEmptyPlan {
			log.Printf("[INFO] Got non-empty plan, as expected:\n\n%s", legacyPlanComparisonString(newState, p.Changes))
		} else {
			return state, fmt.Errorf(
				"After applying this step and refreshing, "+
					"the plan was not empty:\n\n%s", legacyPlanComparisonString(newState, p.Changes))
		}
	}

	// Made it here, but expected a non-empty plan, fail!
	if step.ExpectNonEmptyPlan && empty {
		return state, fmt.Errorf("Expected a non-empty plan, but got an empty plan!")
	}

	// Made it here? Good job test step!
	return state, nil
}

func testStepTaint(state *terraform.State, step TestStep) error {
	for _, p := range step.Taint {
		m := state.RootModule()
		if m == nil {
			return errors.New("no state")
		}
		rs, ok := m.Resources[p]
		if !ok {
			return fmt.Errorf("resource %q not found in state", p)
		}
		log.Printf("[WARN] Test: Explicitly tainting resource %q", p)
		rs.Taint()
	}
	return nil
}
