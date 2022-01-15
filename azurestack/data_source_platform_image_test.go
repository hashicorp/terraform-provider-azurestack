package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataAzureStackPlatformImage_basic(t *testing.T) {
	dataSourceName := "data.azurestack_platform_image.test"
	config := testAccDataAzureStackPlatformImageBasic(testLocation())

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

func testAccDataAzureStackPlatformImageBasic(location string) string {
	return fmt.Sprintf(`
data "azurestack_platform_image" "test" {
  location  = "%s"
  publisher = "Canonical"
  offer     = "UbuntuServer"
  sku       = "16.04-LTS"
}
`, location)
}
