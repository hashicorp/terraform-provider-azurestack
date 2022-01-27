package client

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
)

type Client struct {
	AvailabilitySetsClient          *compute.AvailabilitySetsClient
	DisksClient                     *compute.DisksClient
	VMExtensionImageClient          *compute.VirtualMachineExtensionImagesClient
	VMExtensionClient               *compute.VirtualMachineExtensionsClient
	VMScaleSetClient                *compute.VirtualMachineScaleSetsClient
	VMScaleSetExtensionsClient      *compute.VirtualMachineScaleSetExtensionsClient
	VMScaleSetRollingUpgradesClient *compute.VirtualMachineScaleSetRollingUpgradesClient
	VMScaleSetVMsClient             *compute.VirtualMachineScaleSetVMsClient
	VMClient                        *compute.VirtualMachinesClient
	VMImageClient                   *compute.VirtualMachineImagesClient
}

func NewClient(o *common.ClientOptions) *Client {
	availabilitySetsClient := compute.NewAvailabilitySetsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&availabilitySetsClient.Client, o.ResourceManagerAuthorizer)

	disksClient := compute.NewDisksClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&disksClient.Client, o.ResourceManagerAuthorizer)

	imagesClient := compute.NewImagesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&imagesClient.Client, o.ResourceManagerAuthorizer)

	vmExtensionImageClient := compute.NewVirtualMachineExtensionImagesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&vmExtensionImageClient.Client, o.ResourceManagerAuthorizer)

	vmExtensionClient := compute.NewVirtualMachineExtensionsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&vmExtensionClient.Client, o.ResourceManagerAuthorizer)

	vmImageClient := compute.NewVirtualMachineImagesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&vmImageClient.Client, o.ResourceManagerAuthorizer)

	vmScaleSetClient := compute.NewVirtualMachineScaleSetsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&vmScaleSetClient.Client, o.ResourceManagerAuthorizer)

	vmScaleSetExtensionsClient := compute.NewVirtualMachineScaleSetExtensionsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&vmScaleSetExtensionsClient.Client, o.ResourceManagerAuthorizer)

	vmScaleSetRollingUpgradesClient := compute.NewVirtualMachineScaleSetRollingUpgradesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&vmScaleSetRollingUpgradesClient.Client, o.ResourceManagerAuthorizer)

	vmScaleSetVMsClient := compute.NewVirtualMachineScaleSetVMsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&vmScaleSetVMsClient.Client, o.ResourceManagerAuthorizer)

	vmClient := compute.NewVirtualMachinesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&vmClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		AvailabilitySetsClient:          &availabilitySetsClient,
		DisksClient:                     &disksClient,
		VMExtensionImageClient:          &vmExtensionImageClient,
		VMExtensionClient:               &vmExtensionClient,
		VMScaleSetClient:                &vmScaleSetClient,
		VMScaleSetExtensionsClient:      &vmScaleSetExtensionsClient,
		VMScaleSetRollingUpgradesClient: &vmScaleSetRollingUpgradesClient,
		VMScaleSetVMsClient:             &vmScaleSetVMsClient,
		VMClient:                        &vmClient,
		VMImageClient:                   &vmImageClient,
	}
}
