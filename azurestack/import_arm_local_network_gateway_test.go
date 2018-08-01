package azurestack

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackLocalNetworkGateway_importBasic(t *testing.T) {
	resourceName := "azurestack_local_network_gateway.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLocalNetworkGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLocalNetworkGatewayConfig_basic(rInt, testLocation()),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackLocalNetworkGateway_importBGPSettingsComplete(t *testing.T) {
	resourceName := "azurestack_local_network_gateway.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLocalNetworkGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLocalNetworkGatewayConfig_bgpSettingsComplete(rInt, testLocation()),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
