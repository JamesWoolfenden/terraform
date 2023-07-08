package local

import (
	"testing"

	"github.com/jameswoolfenden/terraform/backend"
	"github.com/jameswoolfenden/terraform/command/arguments"
	"github.com/jameswoolfenden/terraform/command/clistate"
	"github.com/jameswoolfenden/terraform/command/views"
	"github.com/jameswoolfenden/terraform/internal/initwd"
	"github.com/jameswoolfenden/terraform/internal/terminal"
)

func TestLocalContext(t *testing.T) {
	configDir := "./testdata/empty"
	b, cleanup := TestLocal(t)
	defer cleanup()

	_, configLoader, configCleanup := initwd.MustLoadConfigForTests(t, configDir)
	defer configCleanup()

	streams, _ := terminal.StreamsForTesting(t)
	view := views.NewView(streams)
	stateLocker := clistate.NewLocker(0, views.NewStateLocker(arguments.ViewHuman, view))

	op := &backend.Operation{
		ConfigDir:    configDir,
		ConfigLoader: configLoader,
		Workspace:    backend.DefaultStateName,
		StateLocker:  stateLocker,
	}

	_, _, diags := b.Context(op)
	if diags.HasErrors() {
		t.Fatalf("unexpected error: %s", diags.Err().Error())
	}

	// Context() retains a lock on success
	assertBackendStateLocked(t, b)
}

func TestLocalContext_error(t *testing.T) {
	configDir := "./testdata/apply"
	b, cleanup := TestLocal(t)
	defer cleanup()

	_, configLoader, configCleanup := initwd.MustLoadConfigForTests(t, configDir)
	defer configCleanup()

	streams, _ := terminal.StreamsForTesting(t)
	view := views.NewView(streams)
	stateLocker := clistate.NewLocker(0, views.NewStateLocker(arguments.ViewHuman, view))

	op := &backend.Operation{
		ConfigDir:    configDir,
		ConfigLoader: configLoader,
		Workspace:    backend.DefaultStateName,
		StateLocker:  stateLocker,
	}

	_, _, diags := b.Context(op)
	if !diags.HasErrors() {
		t.Fatal("unexpected success")
	}

	// Context() unlocks the state on failure
	assertBackendStateUnlocked(t, b)
}
