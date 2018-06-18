package azurestack

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackSubnet_importBasic(t *testing.T) {
	resourceName := "azurestack_subnet.test"

	ri := acctest.RandInt()
	config := testAccAzureStackSubnet_basic(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Route table not supported yet
func TestAccAzureStackSubnet_importWithRouteTable(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_subnet.test"

	ri := acctest.RandInt()
	config := testAccAzureStackSubnet_routeTable(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackSubnet_importWithNetworkSecurityGroup(t *testing.T) {
	resourceName := "azurestack_subnet.test"

	ri := acctest.RandInt()
	config := testAccAzureStackSubnet_networkSecurityGroup(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
