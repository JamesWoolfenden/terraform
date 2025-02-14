package clistate

import (
	"testing"

	"github.com/jameswoolfenden/terraform/command/arguments"
	"github.com/jameswoolfenden/terraform/command/views"
	"github.com/jameswoolfenden/terraform/internal/terminal"
	"github.com/jameswoolfenden/terraform/states/statemgr"
)

func TestUnlock(t *testing.T) {
	streams, _ := terminal.StreamsForTesting(t)
	view := views.NewView(streams)

	l := NewLocker(0, views.NewStateLocker(arguments.ViewHuman, view))
	l.Lock(statemgr.NewUnlockErrorFull(nil, nil), "test-lock")

	diags := l.Unlock()
	if diags.HasErrors() {
		t.Log(diags.Err().Error())
	} else {
		t.Error("expected error")
	}
}
