package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackLoadBalancerProbe_importBasic(t *testing.T) {
	resourceName := "azurestack_lb_probe.test"

	ri := acctest.RandInt()
	probeName := fmt.Sprintf("probe-%d", ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerProbe_basic(ri, probeName, testLocation()),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// location is deprecated and was never actually used
				ImportStateVerifyIgnore: []string{"location"},
			},
		},
	})
}
