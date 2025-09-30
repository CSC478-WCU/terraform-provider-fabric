package runtime

import "github.com/csc478-wcu/terraform-provider-fabric/internal/services"

type Deps struct {
	Slices        services.SlicesService
	Resources     services.ResourcesService
	DefaultSSHKey string
	Endpoint      string
}
