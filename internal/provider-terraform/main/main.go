package main

import (
	"github.com/jameswoolfenden/terraform/builtin/providers/terraform"
	"github.com/jameswoolfenden/terraform/internal/grpcwrap"
	"github.com/jameswoolfenden/terraform/internal/tfplugin5"
	"github.com/jameswoolfenden/terraform/plugin"
)

func main() {
	// Provide a binary version of the internal terraform provider for testing
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin5.ProviderServer {
			return grpcwrap.Provider(terraform.NewProvider())
		},
	})
}
