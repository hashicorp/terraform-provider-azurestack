package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/blob/blobs"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/blob/containers"

	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"

	// todo switch what we can to github.com/tombuildsstuff/giovanni
	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/storage/mgmt/storage"
	mainStorage "github.com/Azure/azure-sdk-for-go/storage"
)

type Client struct {
	AccountsClient *storage.AccountsClient

	env      azure.Environment
	endpoint string
}

func NewClient(options *common.ClientOptions) *Client {
	accountsClient := storage.NewAccountsClientWithBaseURI(options.ResourceManagerEndpoint, options.SubscriptionId)
	options.ConfigureClient(&accountsClient.Client, options.ResourceManagerAuthorizer)

	client := Client{
		AccountsClient: &accountsClient,
		endpoint:       options.ResourceManagerEndpoint,
		env:            options.Environment,
	}

	return &client
}

var (
	storageKeyCacheMu sync.RWMutex
	storageKeyCache   = make(map[string]string)
)

func (client Client) BlobsClient(ctx context.Context, account accountDetails) (*blobs.Client, error) {
	accountKey, err := account.AccountKey(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("retrieving Account Key: %s", err)
	}

	storageAuth, err := autorest.NewSharedKeyAuthorizer(account.name, *accountKey, autorest.SharedKey)
	if err != nil {
		return nil, fmt.Errorf("building Authorizer: %+v", err)
	}

	blobsClient := blobs.NewWithEnvironment(client.env)
	blobsClient.Client.Authorizer = storageAuth
	return &blobsClient, nil
}

func (client Client) ContainersClient(ctx context.Context, account accountDetails) (*containers.Client, error) {
	accountKey, err := account.AccountKey(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("retrieving Account Key: %s", err)
	}

	storageAuth, err := autorest.NewSharedKeyAuthorizer(account.name, *accountKey, autorest.SharedKey)
	if err != nil {
		return nil, fmt.Errorf("building Authorizer: %+v", err)
	}

	containersClient := containers.NewWithEnvironment(client.env)
	containersClient.Client.Authorizer = storageAuth

	return &containersClient, nil
}

func (client Client) GetKeyForStorageAccount(ctx context.Context, resourceGroupName, storageAccountName string) (string, bool, error) {
	cacheIndex := resourceGroupName + "/" + storageAccountName
	storageKeyCacheMu.RLock()
	key, ok := storageKeyCache[cacheIndex]
	storageKeyCacheMu.RUnlock()

	if ok {
		return key, true, nil
	}

	storageKeyCacheMu.Lock()
	defer storageKeyCacheMu.Unlock()
	key, ok = storageKeyCache[cacheIndex]
	if !ok {
		accountKeys, err := client.AccountsClient.ListKeys(ctx, resourceGroupName, storageAccountName)
		if utils.ResponseWasNotFound(accountKeys.Response) {
			return "", false, nil
		}
		if err != nil {
			// We assume this is a transient error rather than a 404 (which is caught above),  so assume the
			// account still exists.
			return "", true, fmt.Errorf("retrieving keys for storage account %q: %s", storageAccountName, err)
		}

		if accountKeys.Keys == nil {
			return "", false, fmt.Errorf("Nil key returned for storage account %q", storageAccountName)
		}

		keys := *accountKeys.Keys
		if len(keys) == 0 {
			return "", false, fmt.Errorf("No keys returned for storage account %q", storageAccountName)
		}

		keyPtr := keys[0].Value
		if keyPtr == nil {
			return "", false, fmt.Errorf("The first key returned is nil for storage account %q", storageAccountName)
		}

		key = *keyPtr
		storageKeyCache[cacheIndex] = key
	}

	return key, true, nil
}

func (client Client) GetBlobStorageClientForStorageAccount(ctx context.Context, resourceGroupName, storageAccountName string) (*mainStorage.BlobStorageClient, bool, error) {
	key, accountExists, err := client.GetKeyForStorageAccount(ctx, resourceGroupName, storageAccountName)
	if err != nil {
		return nil, accountExists, err
	}
	if !accountExists {
		return nil, false, nil
	}

	storageClient, err := mainStorage.NewClient(storageAccountName, key, client.endpoint,
		"2016-05-31", true)
	if err != nil {
		return nil, true, fmt.Errorf("creating storage client for storage account %q: %s", storageAccountName, err)
	}

	blobClient := storageClient.GetBlobService()
	return &blobClient, true, nil
}
