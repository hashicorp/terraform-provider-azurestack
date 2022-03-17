package clients

import (
	"context"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
	"github.com/hashicorp/terraform-provider-azurestack/internal/features"
	authorization "github.com/hashicorp/terraform-provider-azurestack/internal/services/authorization/client"
	compute "github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/client"
	dns "github.com/hashicorp/terraform-provider-azurestack/internal/services/dns/client"
	keyvault "github.com/hashicorp/terraform-provider-azurestack/internal/services/keyvault/client"
	loadbalancer "github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/client"
	network "github.com/hashicorp/terraform-provider-azurestack/internal/services/network/client"
	resource "github.com/hashicorp/terraform-provider-azurestack/internal/services/resource/client"
	storage "github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/client"
)

type Client struct {
	// StopContext is used for propagating control from Terraform Core (e.g. Ctrl/Cmd+C)
	StopContext context.Context

	Account       *ResourceManagerAccount
	Authorization *authorization.Client
	Compute       *compute.Client
	Dns           *dns.Client
	KeyVault      *keyvault.Client
	LoadBalancer  *loadbalancer.Client
	Network       *network.Client
	Resource      *resource.Client
	Storage       *storage.Client

	Features features.UserFeatures
}

// NOTE: it should be possible for this method to become Private once the top level Client's removed

func (client *Client) Build(ctx context.Context, o *common.ClientOptions) error {
	autorest.Count429AsRetry = false
	// Disable the Azure SDK for Go's validation since it's unhelpful for our use-case
	validation.Disabled = true

	client.StopContext = ctx

	client.Authorization = authorization.NewClient(o)
	client.Compute = compute.NewClient(o)
	client.Dns = dns.NewClient(o)
	client.KeyVault = keyvault.NewClient(o)
	client.LoadBalancer = loadbalancer.NewClient(o)
	client.Network = network.NewClient(o)
	client.Resource = resource.NewClient(o)
	client.Storage = storage.NewClient(o)

	return nil
}
