package main

import (
	"github.com/jameswoolfenden/terraform/internal/grpcwrap"
	simple "github.com/jameswoolfenden/terraform/internal/provider-simple-v6"
	"github.com/jameswoolfenden/terraform/internal/tfplugin6"
	plugin "github.com/jameswoolfenden/terraform/plugin6"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin6.ProviderServer {
			return grpcwrap.Provider6(simple.Provider())
		},
	})
}
