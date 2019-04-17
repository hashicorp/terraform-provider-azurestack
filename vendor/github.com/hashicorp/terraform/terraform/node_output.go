package terraform

import (
	"fmt"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/lang"
)

// NodeApplyableOutput represents an output that is "applyable":
// it is ready to be applied.
type NodeApplyableOutput struct {
	Addr   addrs.AbsOutputValue
	Config *configs.Output // Config is the output in the config
}

var (
	_ GraphNodeSubPath          = (*NodeApplyableOutput)(nil)
	_ RemovableIfNotTargeted    = (*NodeApplyableOutput)(nil)
	_ GraphNodeTargetDownstream = (*NodeApplyableOutput)(nil)
	_ GraphNodeReferenceable    = (*NodeApplyableOutput)(nil)
	_ GraphNodeReferencer       = (*NodeApplyableOutput)(nil)
	_ GraphNodeReferenceOutside = (*NodeApplyableOutput)(nil)
	_ GraphNodeEvalable         = (*NodeApplyableOutput)(nil)
	_ dag.GraphNodeDotter       = (*NodeApplyableOutput)(nil)
)

func (n *NodeApplyableOutput) Name() string {
	return n.Addr.String()
}

// GraphNodeSubPath
func (n *NodeApplyableOutput) Path() addrs.ModuleInstance {
	return n.Addr.Module
}

// RemovableIfNotTargeted
func (n *NodeApplyableOutput) RemoveIfNotTargeted() bool {
	// We need to add this so that this node will be removed if
	// it isn't targeted or a dependency of a target.
	return true
}

// GraphNodeTargetDownstream
func (n *NodeApplyableOutput) TargetDownstream(targetedDeps, untargetedDeps *dag.Set) bool {
	// If any of the direct dependencies of an output are targeted then
	// the output must always be targeted as well, so its value will always
	// be up-to-date at the completion of an apply walk.
	return true
}

func referenceOutsideForOutput(addr addrs.AbsOutputValue) (selfPath, referencePath addrs.ModuleInstance) {

	// Output values have their expressions resolved in the context of the
	// module where they are defined.
	referencePath = addr.Module

	// ...but they are referenced in the context of their calling module.
	selfPath = addr.Module.Parent()

	return // uses named return values

}

// GraphNodeReferenceOutside implementation
func (n *NodeApplyableOutput) ReferenceOutside() (selfPath, referencePath addrs.ModuleInstance) {
	return referenceOutsideForOutput(n.Addr)
}

func referenceableAddrsForOutput(addr addrs.AbsOutputValue) []addrs.Referenceable {
	// An output in the root module can't be referenced at all.
	if addr.Module.IsRoot() {
		return nil
	}

	// Otherwise, we can be referenced via a reference to our output name
	// on the parent module's call, or via a reference to the entire call.
	// e.g. module.foo.bar or just module.foo .
	// Note that our ReferenceOutside method causes these addresses to be
	// relative to the calling module, not the module where the output
	// was declared.
	_, outp := addr.ModuleCallOutput()
	_, call := addr.Module.CallInstance()
	return []addrs.Referenceable{outp, call}

}

// GraphNodeReferenceable
func (n *NodeApplyableOutput) ReferenceableAddrs() []addrs.Referenceable {
	return referenceableAddrsForOutput(n.Addr)
}

func referencesForOutput(c *configs.Output) []*addrs.Reference {
	impRefs, _ := lang.ReferencesInExpr(c.Expr)
	expRefs, _ := lang.References(c.DependsOn)
	l := len(impRefs) + len(expRefs)
	if l == 0 {
		return nil
	}
	refs := make([]*addrs.Reference, 0, l)
	refs = append(refs, impRefs...)
	refs = append(refs, expRefs...)
	return refs

}

// GraphNodeReferencer
func (n *NodeApplyableOutput) References() []*addrs.Reference {
	return appendResourceDestroyReferences(referencesForOutput(n.Config))
}

// GraphNodeEvalable
func (n *NodeApplyableOutput) EvalTree() EvalNode {
	return &EvalSequence{
		Nodes: []EvalNode{
			&EvalOpFilter{
				Ops: []walkOperation{walkRefresh, walkPlan, walkApply, walkValidate, walkDestroy, walkPlanDestroy},
				Node: &EvalWriteOutput{
					Name:          n.Config.Name,
					Sensitive:     n.Config.Sensitive,
					Value:         n.Config.RawConfig,
					ContinueOnErr: true,
				},
			},
			&EvalOpFilter{
				Ops: []walkOperation{walkRefresh, walkPlan, walkApply, walkValidate, walkDestroy, walkPlanDestroy},
				Node: &EvalWriteOutput{
					Name:      n.Config.Name,
					Sensitive: n.Config.Sensitive,
					Expr:      n.Config.Expr,
				},
			},
		},
	}
}

// NodeDestroyableOutput represents an output that is "destroybale":
// its application will remove the output from the state.
type NodeDestroyableOutput struct {
	PathValue []string
	Config    *config.Output // Config is the output in the config
}

func (n *NodeDestroyableOutput) Name() string {
	result := fmt.Sprintf("output.%s (destroy)", n.Config.Name)
	if len(n.PathValue) > 1 {
		result = fmt.Sprintf("%s.%s", modulePrefixStr(n.PathValue), result)
	}

	return result
}

// GraphNodeSubPath
func (n *NodeDestroyableOutput) Path() []string {
	return n.PathValue
}

// RemovableIfNotTargeted
func (n *NodeDestroyableOutput) RemoveIfNotTargeted() bool {
	// We need to add this so that this node will be removed if
	// it isn't targeted or a dependency of a target.
	return true
}

// This will keep the destroy node in the graph if its corresponding output
// node is also in the destroy graph.
func (n *NodeDestroyableOutput) TargetDownstream(targetedDeps, untargetedDeps *dag.Set) bool {
	return true
}

// GraphNodeReferencer
func (n *NodeDestroyableOutput) References() []string {
	var result []string
	result = append(result, n.Config.DependsOn...)
	result = append(result, ReferencesFromConfig(n.Config.RawConfig)...)
	for _, v := range result {
		split := strings.Split(v, "/")
		for i, s := range split {
			split[i] = s + ".destroy"
		}

		result = append(result, strings.Join(split, "/"))
	}

	return result
}

// GraphNodeEvalable
func (n *NodeDestroyableOutput) EvalTree() EvalNode {
	return &EvalDeleteOutput{
		Name: n.Config.Name,
	}
}
