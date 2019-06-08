package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataAZURESTACKPlatformImage_basic(t *testing.T) {
	dataSourceName := "data.azurerm_platform_image.test"
	config := testAccDataAZURESTACKPlatformImageBasic(testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "version"),
					resource.TestCheckResourceAttr(dataSourceName, "publisher", "Canonical"),
					resource.TestCheckResourceAttr(dataSourceName, "offer", "UbuntuServer"),
					resource.TestCheckResourceAttr(dataSourceName, "sku", "16.04-LTS"),
				),
			},
		},
	})
}

func testAccDataAZURESTACKPlatformImageBasic(location string) string {
	return fmt.Sprintf(`
data "azurerm_platform_image" "test" {
  location  = "%s"
  publisher = "Canonical"
  offer     = "UbuntuServer"
  sku       = "16.04-LTS"
}
`, location)
}
