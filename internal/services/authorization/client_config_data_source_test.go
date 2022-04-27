package authorization_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type ClientConfigDataSource struct{}

func TestAccClientConfigDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_client_config", "current")
	clientId := os.Getenv("ARM_CLIENT_ID")
	tenantId := os.Getenv("ARM_TENANT_ID")
	subscriptionId := os.Getenv("ARM_SUBSCRIPTION_ID")
	objectIdRegex := regexp.MustCompile("^[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: ClientConfigDataSource{}.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("client_id").HasValue(clientId),
				check.That(data.ResourceName).Key("tenant_id").HasValue(tenantId),
				check.That(data.ResourceName).Key("subscription_id").HasValue(subscriptionId),
				check.That(data.ResourceName).Key("service_principal_object_id").MatchesRegex(objectIdRegex),
			),
		},
	})
}

func (d ClientConfigDataSource) basic() string {
	return `
data "azurestack_client_config" "current" {
}
`
}
