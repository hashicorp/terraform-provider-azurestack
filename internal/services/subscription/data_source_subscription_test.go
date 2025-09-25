package subscription_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type SubscriptionDataSource struct{}

func TestAccDataSourceSubscription_current(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_subscription", "current")

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: SubscriptionDataSource{}.currentConfig(),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("subscription_id").HasValue(data.Client().SubscriptionID),
				check.That(data.ResourceName).Key("display_name").Exists(),
				check.That(data.ResourceName).Key("tenant_id").Exists(),
				check.That(data.ResourceName).Key("state").HasValue("Enabled"),
			),
		},
	})
}

func TestAccDataSourceSubscription_specific(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_subscription", "specific")

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: SubscriptionDataSource{}.specificConfig(data.Client().SubscriptionID),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("subscription_id").HasValue(data.Client().SubscriptionID),
				check.That(data.ResourceName).Key("display_name").Exists(),
				check.That(data.ResourceName).Key("tenant_id").Exists(),
				check.That(data.ResourceName).Key("location_placement_id").Exists(),
			),
		},
	})
}

func (d SubscriptionDataSource) currentConfig() string {
	return `
provider "azurestack" {
  features {}
}

data "azurestack_subscription" "current" {}
`
}

func (d SubscriptionDataSource) specificConfig(subscriptionId string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
  subscription_id = "%s"
}

data "azurestack_subscription" "specific" {
  subscription_id = "%s"
}
`, subscriptionId, subscriptionId)
}
