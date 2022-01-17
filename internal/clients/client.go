package clients

import (
	"context"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
	"github.com/hashicorp/terraform-provider-azurestack/internal/features"
	authorization "github.com/hashicorp/terraform-provider-azurestack/internal/services/authorization/client"
	dns "github.com/hashicorp/terraform-provider-azurestack/internal/services/dns/client"
	resource "github.com/hashicorp/terraform-provider-azurestack/internal/services/resource/client"
)

type Client struct {
	// StopContext is used for propagating control from Terraform Core (e.g. Ctrl/Cmd+C)
	StopContext context.Context

	Account       *ResourceManagerAccount
	Authorization *authorization.Client
	Dns           *dns.Client
	Features      features.UserFeatures

	Resource *resource.Client
}

// NOTE: it should be possible for this method to become Private once the top level Client's removed

func (client *Client) Build(ctx context.Context, o *common.ClientOptions) error {
	autorest.Count429AsRetry = false
	// Disable the Azure SDK for Go's validation since it's unhelpful for our use-case
	validation.Disabled = true

	client.StopContext = ctx

	client.Authorization = authorization.NewClient(o)
	client.Dns = dns.NewClient(o)
	client.Resource = resource.NewClient(o)

	return nil
}
