package terraform

import (
	backendInit "github.com/jameswoolfenden/terraform/backend/init"
)

func init() {
	// Initialize the backends
	backendInit.Init(nil)
}
