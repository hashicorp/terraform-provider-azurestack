package subscription_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type SubscriptionsDataSource struct{}

func TestAccDataSourceSubscriptions_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_subscriptions", "current")

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: SubscriptionsDataSource{}.basic(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("subscriptions.0.id").Exists(),
				check.That(data.ResourceName).Key("subscriptions.0.subscription_id").Exists(),
				check.That(data.ResourceName).Key("subscriptions.0.display_name").Exists(),
				check.That(data.ResourceName).Key("subscriptions.0.tenant_id").Exists(),
				check.That(data.ResourceName).Key("subscriptions.0.state").Exists(),
			),
		},
	})
}

func (d SubscriptionsDataSource) basic() string {
	return `
provider "azurestack" {
  features {}
}

data "azurestack_subscriptions" "current" {}
`
}
