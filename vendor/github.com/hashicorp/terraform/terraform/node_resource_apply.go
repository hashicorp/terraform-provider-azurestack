package terraform

import (
	"log"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/lang"
)

// NodeApplyableResource represents a resource that is "applyable":
// it may need to have its record in the state adjusted to match configuration.
//
// Unlike in the plan walk, this resource node does not DynamicExpand. Instead,
// it should be inserted into the same graph as any instances of the nodes
// with dependency edges ensuring that the resource is evaluated before any
// of its instances, which will turn ensure that the whole-resource record
// in the state is suitably prepared to receive any updates to instances.
type NodeApplyableResource struct {
	*NodeAbstractResource
}

var (
	_ GraphNodeResource             = (*NodeApplyableResource)(nil)
	_ GraphNodeEvalable             = (*NodeApplyableResource)(nil)
	_ GraphNodeProviderConsumer     = (*NodeApplyableResource)(nil)
	_ GraphNodeAttachResourceConfig = (*NodeApplyableResource)(nil)
	_ GraphNodeReferencer           = (*NodeApplyableResource)(nil)
)

func (n *NodeApplyableResource) Name() string {
	return n.NodeAbstractResource.Name() + " (prepare state)"
}

func (n *NodeApplyableResource) References() []*addrs.Reference {
	if n.Config == nil {
		log.Printf("[WARN] NodeApplyableResource %q: no configuration, so can't determine References", dag.VertexName(n))
		return nil
	}

	var result []*addrs.Reference

	// Since this node type only updates resource-level metadata, we only
	// need to worry about the parts of the configuration that affect
	// our "each mode": the count and for_each meta-arguments.
	refs, _ := lang.ReferencesInExpr(n.Config.Count)
	result = append(result, refs...)
	refs, _ = lang.ReferencesInExpr(n.Config.ForEach)
	result = append(result, refs...)

	return result
}

// GraphNodeEvalable
func (n *NodeApplyableResource) EvalTree() EvalNode {
	addr := n.ResourceAddr()
	config := n.Config
	providerAddr := n.ResolvedProvider

	if config == nil {
		// Nothing to do, then.
		log.Printf("[TRACE] NodeApplyableResource: no configuration present for %s", addr)
		return &EvalNoop{}
	}

	return &EvalWriteResourceState{
		Addr:         addr.Resource,
		Config:       config,
		ProviderAddr: providerAddr,
	}
}

func (n *NodeApplyableResource) evalTreeDataResource(
	stateId string, info *InstanceInfo,
	resource *Resource, stateDeps []string) EvalNode {
	var provider ResourceProvider
	var config *ResourceConfig
	var diff *InstanceDiff
	var state *InstanceState

	return &EvalSequence{
		Nodes: []EvalNode{
			// Build the instance info
			&EvalInstanceInfo{
				Info: info,
			},

			// Get the saved diff for apply
			&EvalReadDiff{
				Name: stateId,
				Diff: &diff,
			},

			// Stop here if we don't actually have a diff
			&EvalIf{
				If: func(ctx EvalContext) (bool, error) {
					if diff == nil {
						return true, EvalEarlyExitError{}
					}

					if diff.GetAttributesLen() == 0 {
						return true, EvalEarlyExitError{}
					}

					return true, nil
				},
				Then: EvalNoop{},
			},

			// Normally we interpolate count as a preparation step before
			// a DynamicExpand, but an apply graph has pre-expanded nodes
			// and so the count would otherwise never be interpolated.
			//
			// This is redundant when there are multiple instances created
			// from the same config (count > 1) but harmless since the
			// underlying structures have mutexes to make this concurrency-safe.
			//
			// In most cases this isn't actually needed because we dealt with
			// all of the counts during the plan walk, but we do it here
			// for completeness because other code assumes that the
			// final count is always available during interpolation.
			//
			// Here we are just populating the interpolated value in-place
			// inside this RawConfig object, like we would in
			// NodeAbstractCountResource.
			&EvalInterpolate{
				Config:        n.Config.RawCount,
				ContinueOnErr: true,
			},

			// We need to re-interpolate the config here, rather than
			// just using the diff's values directly, because we've
			// potentially learned more variable values during the
			// apply pass that weren't known when the diff was produced.
			&EvalInterpolate{
				Config:   n.Config.RawConfig.Copy(),
				Resource: resource,
				Output:   &config,
			},

			&EvalGetProvider{
				Name:   n.ResolvedProvider,
				Output: &provider,
			},

			// Make a new diff with our newly-interpolated config.
			&EvalReadDataDiff{
				Info:     info,
				Config:   &config,
				Previous: &diff,
				Provider: &provider,
				Output:   &diff,
			},

			&EvalReadDataApply{
				Info:     info,
				Diff:     &diff,
				Provider: &provider,
				Output:   &state,
			},

			&EvalWriteState{
				Name:         stateId,
				ResourceType: n.Config.Type,
				Provider:     n.ResolvedProvider,
				Dependencies: stateDeps,
				State:        &state,
			},

			// Clear the diff now that we've applied it, so
			// later nodes won't see a diff that's now a no-op.
			&EvalWriteDiff{
				Name: stateId,
				Diff: nil,
			},

			&EvalUpdateStateHook{},
		},
	}
}

func (n *NodeApplyableResource) evalTreeManagedResource(
	stateId string, info *InstanceInfo,
	resource *Resource, stateDeps []string) EvalNode {
	// Declare a bunch of variables that are used for state during
	// evaluation. Most of this are written to by-address below.
	var provider ResourceProvider
	var diff, diffApply *InstanceDiff
	var state *InstanceState
	var resourceConfig *ResourceConfig
	var err error
	var createNew bool
	var createBeforeDestroyEnabled bool

	return &EvalSequence{
		Nodes: []EvalNode{
			// Build the instance info
			&EvalInstanceInfo{
				Info: info,
			},

			// Get the saved diff for apply
			&EvalReadDiff{
				Name: stateId,
				Diff: &diffApply,
			},

			// We don't want to do any destroys
			&EvalIf{
				If: func(ctx EvalContext) (bool, error) {
					if diffApply == nil {
						return true, EvalEarlyExitError{}
					}

					if diffApply.GetDestroy() && diffApply.GetAttributesLen() == 0 {
						return true, EvalEarlyExitError{}
					}

					diffApply.SetDestroy(false)
					return true, nil
				},
				Then: EvalNoop{},
			},

			&EvalIf{
				If: func(ctx EvalContext) (bool, error) {
					destroy := false
					if diffApply != nil {
						destroy = diffApply.GetDestroy() || diffApply.RequiresNew()
					}

					createBeforeDestroyEnabled =
						n.Config.Lifecycle.CreateBeforeDestroy &&
							destroy

					return createBeforeDestroyEnabled, nil
				},
				Then: &EvalDeposeState{
					Name: stateId,
				},
			},

			// Normally we interpolate count as a preparation step before
			// a DynamicExpand, but an apply graph has pre-expanded nodes
			// and so the count would otherwise never be interpolated.
			//
			// This is redundant when there are multiple instances created
			// from the same config (count > 1) but harmless since the
			// underlying structures have mutexes to make this concurrency-safe.
			//
			// In most cases this isn't actually needed because we dealt with
			// all of the counts during the plan walk, but we need to do this
			// in order to support interpolation of resource counts from
			// apply-time-interpolated expressions, such as those in
			// "provisioner" blocks.
			//
			// Here we are just populating the interpolated value in-place
			// inside this RawConfig object, like we would in
			// NodeAbstractCountResource.
			&EvalInterpolate{
				Config:        n.Config.RawCount,
				ContinueOnErr: true,
			},

			&EvalInterpolate{
				Config:   n.Config.RawConfig.Copy(),
				Resource: resource,
				Output:   &resourceConfig,
			},
			&EvalGetProvider{
				Name:   n.ResolvedProvider,
				Output: &provider,
			},
			&EvalReadState{
				Name:   stateId,
				Output: &state,
			},
			// Re-run validation to catch any errors we missed, e.g. type
			// mismatches on computed values.
			&EvalValidateResource{
				Provider:       &provider,
				Config:         &resourceConfig,
				ResourceName:   n.Config.Name,
				ResourceType:   n.Config.Type,
				ResourceMode:   n.Config.Mode,
				IgnoreWarnings: true,
			},
			&EvalDiff{
				Info:       info,
				Config:     &resourceConfig,
				Resource:   n.Config,
				Provider:   &provider,
				Diff:       &diffApply,
				State:      &state,
				OutputDiff: &diffApply,
			},

			// Get the saved diff
			&EvalReadDiff{
				Name: stateId,
				Diff: &diff,
			},

			// Compare the diffs
			&EvalCompareDiff{
				Info: info,
				One:  &diff,
				Two:  &diffApply,
			},

			&EvalGetProvider{
				Name:   n.ResolvedProvider,
				Output: &provider,
			},
			&EvalReadState{
				Name:   stateId,
				Output: &state,
			},
			// Call pre-apply hook
			&EvalApplyPre{
				Info:  info,
				State: &state,
				Diff:  &diffApply,
			},
			&EvalApply{
				Info:      info,
				State:     &state,
				Diff:      &diffApply,
				Provider:  &provider,
				Output:    &state,
				Error:     &err,
				CreateNew: &createNew,
			},
			&EvalWriteState{
				Name:         stateId,
				ResourceType: n.Config.Type,
				Provider:     n.ResolvedProvider,
				Dependencies: stateDeps,
				State:        &state,
			},
			&EvalApplyProvisioners{
				Info:           info,
				State:          &state,
				Resource:       n.Config,
				InterpResource: resource,
				CreateNew:      &createNew,
				Error:          &err,
				When:           config.ProvisionerWhenCreate,
			},
			&EvalIf{
				If: func(ctx EvalContext) (bool, error) {
					return createBeforeDestroyEnabled && err != nil, nil
				},
				Then: &EvalUndeposeState{
					Name:  stateId,
					State: &state,
				},
				Else: &EvalWriteState{
					Name:         stateId,
					ResourceType: n.Config.Type,
					Provider:     n.ResolvedProvider,
					Dependencies: stateDeps,
					State:        &state,
				},
			},

			// We clear the diff out here so that future nodes
			// don't see a diff that is already complete. There
			// is no longer a diff!
			&EvalWriteDiff{
				Name: stateId,
				Diff: nil,
			},

			&EvalApplyPost{
				Info:  info,
				State: &state,
				Error: &err,
			},
			&EvalUpdateStateHook{},
		},
	}
}
