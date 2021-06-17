package client

import (
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/common"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/dataprotection/legacysdk/dataprotection"
)

type Client struct {
	BackupVaultClient *dataprotection.BackupVaultsClient
}

func NewClient(o *common.ClientOptions) *Client {
	backupVaultClient := dataprotection.NewBackupVaultsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&backupVaultClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		BackupVaultClient: &backupVaultClient,
	}
}
