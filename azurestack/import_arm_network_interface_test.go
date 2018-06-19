package azurestack

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackNetworkInterface_importBasic(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_basic(rInt, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackNetworkInterface_importIPForwarding(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_ipForwarding(rInt, testLocation()),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enable_accelerated_networking"},
			},
		},
	})
}

func TestAccAzureStackNetworkInterface_importWithTags(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_withTags(rInt, testLocation()),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enable_accelerated_networking"},
			},
		},
	})
}

// Load Balancer not yet supported
func TestAccAzureStackNetworkInterface_importMultipleLoadBalancers(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test1"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_multipleLoadBalancers(rInt, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// App gateway not supported
func TestAccAzureStackNetworkInterface_importApplicationGateway(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_applicationGatewayBackendPool(rInt, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// public IP not yet supported
func TestAccAzureStackNetworkInterface_importPublicIP(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_publicIP(rInt, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// App Security Group not supported
func TestAccAzureStackNetworkInterface_importApplicationSecurityGroup(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_applicationSecurityGroup(rInt, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
