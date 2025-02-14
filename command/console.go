package command

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/jameswoolfenden/terraform/addrs"
	"github.com/jameswoolfenden/terraform/backend"
	"github.com/jameswoolfenden/terraform/internal/helper/wrappedstreams"
	"github.com/jameswoolfenden/terraform/repl"
	"github.com/jameswoolfenden/terraform/tfdiags"

	"github.com/mitchellh/cli"
)

// ConsoleCommand is a Command implementation that applies a Terraform
// configuration and actually builds or changes infrastructure.
type ConsoleCommand struct {
	Meta
}

func (c *ConsoleCommand) Run(args []string) int {
	args = c.Meta.process(args)
	cmdFlags := c.Meta.extendedFlagSet("console")
	cmdFlags.StringVar(&c.Meta.statePath, "state", DefaultStateFilename, "path")
	cmdFlags.Usage = func() { c.Ui.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing command line flags: %s\n", err.Error()))
		return 1
	}

	configPath, err := ModulePath(cmdFlags.Args())
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	configPath = c.Meta.normalizePath(configPath)

	// Check for user-supplied plugin path
	if c.pluginPath, err = c.loadPluginPath(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error loading plugin path: %s", err))
		return 1
	}

	var diags tfdiags.Diagnostics

	backendConfig, backendDiags := c.loadBackendConfig(configPath)
	diags = diags.Append(backendDiags)
	if diags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// Load the backend
	b, backendDiags := c.Backend(&BackendOpts{
		Config: backendConfig,
	})
	diags = diags.Append(backendDiags)
	if backendDiags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// We require a local backend
	local, ok := b.(backend.Local)
	if !ok {
		c.showDiagnostics(diags) // in case of any warnings in here
		c.Ui.Error(ErrUnsupportedLocalOp)
		return 1
	}

	// This is a read-only command
	c.ignoreRemoteBackendVersionConflict(b)

	// Build the operation
	opReq := c.Operation(b)
	opReq.ConfigDir = configPath
	opReq.ConfigLoader, err = c.initConfigLoader()
	opReq.AllowUnsetVariables = true // we'll just evaluate them as unknown
	if err != nil {
		diags = diags.Append(err)
		c.showDiagnostics(diags)
		return 1
	}

	{
		var moreDiags tfdiags.Diagnostics
		opReq.Variables, moreDiags = c.collectVariableValues()
		diags = diags.Append(moreDiags)
		if moreDiags.HasErrors() {
			c.showDiagnostics(diags)
			return 1
		}
	}

	// Get the context
	ctx, _, ctxDiags := local.Context(opReq)
	diags = diags.Append(ctxDiags)
	if ctxDiags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// Successfully creating the context can result in a lock, so ensure we release it
	defer func() {
		diags := opReq.StateLocker.Unlock()
		if diags.HasErrors() {
			c.showDiagnostics(diags)
		}
	}()

	// Set up the UI so we can output directly to stdout
	ui := &cli.BasicUi{
		Writer:      wrappedstreams.Stdout(),
		ErrorWriter: wrappedstreams.Stderr(),
	}

	// Before we can evaluate expressions, we must compute and populate any
	// derived values (input variables, local values, output values)
	// that are not stored in the persistent state.
	scope, scopeDiags := ctx.Eval(addrs.RootModuleInstance)
	diags = diags.Append(scopeDiags)
	if scope == nil {
		// scope is nil if there are errors so bad that we can't even build a scope.
		// Otherwise, we'll try to eval anyway.
		c.showDiagnostics(diags)
		return 1
	}
	if diags.HasErrors() {
		diags = diags.Append(tfdiags.SimpleWarning("Due to the problems above, some expressions may produce unexpected results."))
	}

	// Before we become interactive we'll show any diagnostics we encountered
	// during initialization, and then afterwards the driver will manage any
	// further diagnostics itself.
	c.showDiagnostics(diags)

	// IO Loop
	session := &repl.Session{
		Scope: scope,
	}

	// Determine if stdin is a pipe. If so, we evaluate directly.
	if c.StdinPiped() {
		return c.modePiped(session, ui)
	}

	return c.modeInteractive(session, ui)
}

func (c *ConsoleCommand) modePiped(session *repl.Session, ui cli.Ui) int {
	var lastResult string
	scanner := bufio.NewScanner(wrappedstreams.Stdin())
	for scanner.Scan() {
		result, exit, diags := session.Handle(strings.TrimSpace(scanner.Text()))
		if diags.HasErrors() {
			// In piped mode we'll exit immediately on error.
			c.showDiagnostics(diags)
			return 1
		}
		if exit {
			return 0
		}

		// Store the last result
		lastResult = result
	}

	// Output the final result
	ui.Output(lastResult)

	return 0
}

func (c *ConsoleCommand) Help() string {
	helpText := `
Usage: terraform [global options] console [options]

  Starts an interactive console for experimenting with Terraform
  interpolations.

  This will open an interactive console that you can use to type
  interpolations into and inspect their values. This command loads the
  current state. This lets you explore and test interpolations before
  using them in future configurations.

  This command will never modify your state.

Options:

  -state=path       Legacy option for the local backend only. See the local
                    backend's documentation for more information.

  -var 'foo=bar'    Set a variable in the Terraform configuration. This
                    flag can be set multiple times.

  -var-file=foo     Set variables in the Terraform configuration from
                    a file. If "terraform.tfvars" or any ".auto.tfvars"
                    files are present, they will be automatically loaded.
`
	return strings.TrimSpace(helpText)
}

func (c *ConsoleCommand) Synopsis() string {
	return "Try Terraform expressions at an interactive command prompt"
}
