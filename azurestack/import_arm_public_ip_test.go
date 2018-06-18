package azurestack

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackPublicIpStatic_importBasic(t *testing.T) {
	resourceName := "azurestack_public_ip.test"

	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPStatic_basic(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
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

// Zones not supported in the profile, skipping
func TestAccAzureStackPublicIpStatic_importBasic_withZone(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_public_ip.test"

	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPStatic_basic_withZone(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
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

func TestAccAzureStackPublicIpStatic_importIdError(t *testing.T) {
	resourceName := "azurestack_public_ip.test"

	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPStatic_basic(ri, testLocation())
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/publicIPAdresses/acctestpublicip-%d", os.Getenv("ARM_SUBSCRIPTION_ID"), ri, ri),
				ExpectError:       regexp.MustCompile("Error parsing supplied resource id."),
			},
		},
	})
}
