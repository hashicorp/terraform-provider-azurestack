package azurestack

import (
	hybridcompute "github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/network/mgmt/network"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/storage/mgmt/storage"
)

// ArmClient contains the handles to all the specific Azure Resource Manager
// resource classes' respective clients.
type ArmClient struct {
	vmClient      hybridcompute.VirtualMachinesClient
	networkClient network.VirtualNetworksClient
	storageClient storage.AccountsClient
}
