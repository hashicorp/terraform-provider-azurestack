package azurestack

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackAvailabilitySet_importBasic(t *testing.T) {
	resourceName := "azurestack_availability_set.test"

	ri := acctest.RandInt()
	config := testAccAzureStackAvailabilitySet_basic(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
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

func TestAccAzureStackAvailabilitySet_importWithTags(t *testing.T) {
	resourceName := "azurestack_availability_set.test"

	ri := acctest.RandInt()
	config := testAccAzureStackAvailabilitySet_withTags(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
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

func TestAccAzureStackAvailabilitySet_importWithDomainCounts(t *testing.T) {
	resourceName := "azurestack_availability_set.test"

	ri := acctest.RandInt()
	config := testAccAzureStackAvailabilitySet_withDomainCounts(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
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

// Managed is not supported in the profile, skipping
func TestAccAzureStackAvailabilitySet_importManaged(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_availability_set.test"

	ri := acctest.RandInt()
	config := testAccAzureStackAvailabilitySet_managed(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
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
