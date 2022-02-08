package compute_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type LinuxVirtualMachineResource struct{}

func (t LinuxVirtualMachineResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualMachineID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.VMClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Linux Virtual Machine %q", id)
	}

	return utils.Bool(resp.ID != nil), nil
}
