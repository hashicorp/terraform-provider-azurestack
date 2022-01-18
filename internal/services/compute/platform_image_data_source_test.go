package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type PlatformImageDataSource struct{}

func TestAccPlatformImageDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_platform_image", "test")
	r := PlatformImageDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("version").Exists(),
				check.That(data.ResourceName).Key("publisher").HasValue("Canonical"),
				check.That(data.ResourceName).Key("offer").HasValue("UbuntuServer"),
				check.That(data.ResourceName).Key("sku").HasValue("16.04-LTS"),
			),
		},
	})
}

func TestAccPlatformImageDataSource_withVersion(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_platform_image", "test")
	r := PlatformImageDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.withVersion(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("version").Exists(),
				check.That(data.ResourceName).Key("publisher").HasValue("Canonical"),
				check.That(data.ResourceName).Key("offer").HasValue("UbuntuServer"),
				check.That(data.ResourceName).Key("sku").HasValue("16.04-LTS"),
				check.That(data.ResourceName).Key("version").HasValue("16.04.201811010"),
			),
		},
	})
}

func (PlatformImageDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_platform_image" "test" {
  location  = "%s"
  publisher = "Canonical"
  offer     = "UbuntuServer"
  sku       = "16.04-LTS"
}
`, data.Locations.Primary)
}

func (PlatformImageDataSource) withVersion(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_platform_image" "test" {
  location  = "%s"
  publisher = "Canonical"
  offer     = "UbuntuServer"
  sku       = "16.04-LTS"
  version   = "16.04.201811010"
}
`, data.Locations.Primary)
}
