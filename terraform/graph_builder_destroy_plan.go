package terraform

import (
	"github.com/jameswoolfenden/terraform/addrs"
	"github.com/jameswoolfenden/terraform/configs"
	"github.com/jameswoolfenden/terraform/dag"
	"github.com/jameswoolfenden/terraform/states"
	"github.com/jameswoolfenden/terraform/tfdiags"
)

// DestroyPlanGraphBuilder implements GraphBuilder and is responsible for
// planning a pure-destroy.
//
// Planning a pure destroy operation is simple because we can ignore most
// ordering configuration and simply reverse the state. This graph mainly
// exists for targeting, because we need to walk the destroy dependencies to
// ensure we plan the required resources. Without the requirement for
// targeting, the plan could theoretically be created directly from the state.
type DestroyPlanGraphBuilder struct {
	// Config is the configuration tree to build the plan from.
	Config *configs.Config

	// State is the current state
	State *states.State

	// Components is a factory for the plug-in components (providers and
	// provisioners) available for use.
	Components contextComponentFactory

	// Schemas is the repository of schemas we will draw from to analyse
	// the configuration.
	Schemas *Schemas

	// Targets are resources to target
	Targets []addrs.Targetable

	// Validate will do structural validation of the graph.
	Validate bool
}

// See GraphBuilder
func (b *DestroyPlanGraphBuilder) Build(path addrs.ModuleInstance) (*Graph, tfdiags.Diagnostics) {
	return (&BasicGraphBuilder{
		Steps:    b.Steps(),
		Validate: b.Validate,
		Name:     "DestroyPlanGraphBuilder",
	}).Build(path)
}

// See GraphBuilder
func (b *DestroyPlanGraphBuilder) Steps() []GraphTransformer {
	concreteResourceInstance := func(a *NodeAbstractResourceInstance) dag.Vertex {
		return &NodePlanDestroyableResourceInstance{
			NodeAbstractResourceInstance: a,
		}
	}
	concreteResourceInstanceDeposed := func(a *NodeAbstractResourceInstance, key states.DeposedKey) dag.Vertex {
		return &NodePlanDeposedResourceInstanceObject{
			NodeAbstractResourceInstance: a,
			DeposedKey:                   key,
		}
	}

	concreteProvider := func(a *NodeAbstractProvider) dag.Vertex {
		return &NodeApplyableProvider{
			NodeAbstractProvider: a,
		}
	}

	steps := []GraphTransformer{
		// Creates nodes for the resource instances tracked in the state.
		&StateTransformer{
			ConcreteCurrent: concreteResourceInstance,
			ConcreteDeposed: concreteResourceInstanceDeposed,
			State:           b.State,
		},

		// Create the delete changes for root module outputs.
		&OutputTransformer{
			Config:  b.Config,
			Destroy: true,
		},

		// Attach the state
		&AttachStateTransformer{State: b.State},

		// Attach the configuration to any resources
		&AttachResourceConfigTransformer{Config: b.Config},

		TransformProviders(b.Components.ResourceProviders(), concreteProvider, b.Config),

		// Destruction ordering. We require this only so that
		// targeting below will prune the correct things.
		&DestroyEdgeTransformer{
			Config:  b.Config,
			State:   b.State,
			Schemas: b.Schemas,
		},

		&TargetsTransformer{Targets: b.Targets},

		// Close opened plugin connections
		&CloseProviderTransformer{},

		// Close the root module
		&CloseRootModuleTransformer{},

		&TransitiveReductionTransformer{},
	}

	return steps
}
