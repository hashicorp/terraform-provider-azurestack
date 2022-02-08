package compute_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type LinuxVirtualMachineScaleSetResource struct{}

func (r LinuxVirtualMachineScaleSetResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualMachineScaleSetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.VMScaleSetClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Linux Virtual Machine Scale Set %q: %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}
