package views

import (
	"testing"

	"github.com/jameswoolfenden/terraform/addrs"
	"github.com/jameswoolfenden/terraform/command/arguments"
	"github.com/jameswoolfenden/terraform/configs/configschema"
	"github.com/jameswoolfenden/terraform/internal/terminal"
	"github.com/jameswoolfenden/terraform/plans"
	"github.com/jameswoolfenden/terraform/providers"
	"github.com/jameswoolfenden/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

// Ensure that the correct view type and in-automation settings propagate to the
// Operation view.
func TestPlanHuman_operation(t *testing.T) {
	streams, done := terminal.StreamsForTesting(t)
	defer done(t)
	v := NewPlan(arguments.ViewHuman, true, NewView(streams)).Operation()
	if hv, ok := v.(*OperationHuman); !ok {
		t.Fatalf("unexpected return type %t", v)
	} else if hv.inAutomation != true {
		t.Fatalf("unexpected inAutomation value on Operation view")
	}
}

// Verify that Hooks includes a UI hook
func TestPlanHuman_hooks(t *testing.T) {
	streams, done := terminal.StreamsForTesting(t)
	defer done(t)
	v := NewPlan(arguments.ViewHuman, true, NewView(streams))
	hooks := v.Hooks()

	var uiHook *UiHook
	for _, hook := range hooks {
		if ch, ok := hook.(*UiHook); ok {
			uiHook = ch
		}
	}
	if uiHook == nil {
		t.Fatalf("expected Hooks to include a UiHook: %#v", hooks)
	}
}

// Helper functions to build a trivial test plan, to exercise the plan
// renderer.
func testPlan(t *testing.T) *plans.Plan {
	t.Helper()

	plannedVal := cty.ObjectVal(map[string]cty.Value{
		"id":  cty.UnknownVal(cty.String),
		"foo": cty.StringVal("bar"),
	})
	priorValRaw, err := plans.NewDynamicValue(cty.NullVal(plannedVal.Type()), plannedVal.Type())
	if err != nil {
		t.Fatal(err)
	}
	plannedValRaw, err := plans.NewDynamicValue(plannedVal, plannedVal.Type())
	if err != nil {
		t.Fatal(err)
	}

	changes := plans.NewChanges()
	changes.SyncWrapper().AppendResourceInstanceChange(&plans.ResourceInstanceChangeSrc{
		Addr: addrs.Resource{
			Mode: addrs.ManagedResourceMode,
			Type: "test_resource",
			Name: "foo",
		}.Instance(addrs.NoKey).Absolute(addrs.RootModuleInstance),
		ProviderAddr: addrs.AbsProviderConfig{
			Provider: addrs.NewDefaultProvider("test"),
			Module:   addrs.RootModule,
		},
		ChangeSrc: plans.ChangeSrc{
			Action: plans.Create,
			Before: priorValRaw,
			After:  plannedValRaw,
		},
	})

	return &plans.Plan{
		Changes: changes,
	}
}

func testSchemas() *terraform.Schemas {
	provider := testProvider()
	return &terraform.Schemas{
		Providers: map[addrs.Provider]*terraform.ProviderSchema{
			addrs.NewDefaultProvider("test"): provider.ProviderSchema(),
		},
	}
}

func testProvider() *terraform.MockProvider {
	p := new(terraform.MockProvider)
	p.ReadResourceFn = func(req providers.ReadResourceRequest) providers.ReadResourceResponse {
		return providers.ReadResourceResponse{NewState: req.PriorState}
	}

	p.GetProviderSchemaResponse = testProviderSchema()

	return p
}

func testProviderSchema() *providers.GetProviderSchemaResponse {
	return &providers.GetProviderSchemaResponse{
		Provider: providers.Schema{
			Block: &configschema.Block{},
		},
		ResourceTypes: map[string]providers.Schema{
			"test_resource": {
				Block: &configschema.Block{
					Attributes: map[string]*configschema.Attribute{
						"id":  {Type: cty.String, Computed: true},
						"foo": {Type: cty.String, Optional: true},
					},
				},
			},
		},
	}
}
