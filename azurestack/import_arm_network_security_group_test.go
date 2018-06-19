package azurestack

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackNetworkSecurityGroup_importBasic(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_basic(rInt, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccazureStackNetworkSecurityGroup_importSingleRule(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_singleRule(rInt, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackNetworkSecurityGroup_importMultipleRules(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_anotherRule(rInt, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
