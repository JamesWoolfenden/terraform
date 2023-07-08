package main

import (
	localexec "github.com/jameswoolfenden/terraform/builtin/provisioners/local-exec"
	"github.com/jameswoolfenden/terraform/internal/grpcwrap"
	"github.com/jameswoolfenden/terraform/internal/tfplugin5"
	"github.com/jameswoolfenden/terraform/plugin"
)

func main() {
	// Provide a binary version of the internal terraform provider for testing
	plugin.Serve(&plugin.ServeOpts{
		GRPCProvisionerFunc: func() tfplugin5.ProvisionerServer {
			return grpcwrap.Provisioner(localexec.New())
		},
	})
}
