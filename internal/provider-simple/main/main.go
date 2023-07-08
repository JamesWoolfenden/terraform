package main

import (
	"github.com/jameswoolfenden/terraform/internal/grpcwrap"
	simple "github.com/jameswoolfenden/terraform/internal/provider-simple"
	"github.com/jameswoolfenden/terraform/internal/tfplugin5"
	"github.com/jameswoolfenden/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin5.ProviderServer {
			return grpcwrap.Provider(simple.Provider())
		},
	})
}
