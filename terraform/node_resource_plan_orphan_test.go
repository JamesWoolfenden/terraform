package terraform

import (
	"testing"

	"github.com/jameswoolfenden/terraform/addrs"
	"github.com/jameswoolfenden/terraform/configs/configschema"
	"github.com/jameswoolfenden/terraform/instances"
	"github.com/jameswoolfenden/terraform/plans"
	"github.com/jameswoolfenden/terraform/states"
)

func TestNodeResourcePlanOrphanExecute(t *testing.T) {
	state := states.NewState()
	state.Module(addrs.RootModuleInstance).SetResourceInstanceCurrent(
		addrs.Resource{
			Mode: addrs.ManagedResourceMode,
			Type: "test_object",
			Name: "foo",
		}.Instance(addrs.NoKey),
		&states.ResourceInstanceObjectSrc{
			AttrsFlat: map[string]string{
				"test_string": "foo",
			},
			Status: states.ObjectReady,
		},
		addrs.AbsProviderConfig{
			Provider: addrs.NewDefaultProvider("test"),
			Module:   addrs.RootModule,
		},
	)

	p := simpleMockProvider()
	ctx := &MockEvalContext{
		StateState:               state.SyncWrapper(),
		RefreshStateState:        state.SyncWrapper(),
		InstanceExpanderExpander: instances.NewExpander(),
		ProviderProvider:         p,
		ProviderSchemaSchema: &ProviderSchema{
			ResourceTypes: map[string]*configschema.Block{
				"test_object": simpleTestSchema(),
			},
		},
		ChangesChanges: plans.NewChanges().SyncWrapper(),
	}

	node := NodePlannableResourceInstanceOrphan{
		NodeAbstractResourceInstance: &NodeAbstractResourceInstance{
			NodeAbstractResource: NodeAbstractResource{
				ResolvedProvider: addrs.AbsProviderConfig{
					Provider: addrs.NewDefaultProvider("test"),
					Module:   addrs.RootModule,
				},
			},
			Addr: mustResourceInstanceAddr("test_object.foo"),
		},
	}
	diags := node.Execute(ctx, walkApply)
	if diags.HasErrors() {
		t.Fatalf("unexpected error: %s", diags.Err())
	}
	if !state.Empty() {
		t.Fatalf("expected empty state, got %s", state.String())
	}
}
