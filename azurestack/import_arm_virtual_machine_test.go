package azurestack

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackVirtualMachine_importBasic(t *testing.T) {
	resourceName := "azurestack_virtual_machine.test"

	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_basicLinuxMachine(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_data_disks_on_termination",
					"delete_os_disk_on_termination",
				},
			},
		},
	})
}

// func TestAccAzureStackVirtualMachine_importBasic_withZone(t *testing.T) {
// 	resourceName := "azurestack_virtual_machine.test"
//
// 	ri := acctest.RandInt()
// 	config := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_implicit_withZone(ri, testLocation())
//
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: config,
// 			},
//
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 				ImportStateVerifyIgnore: []string{
// 					"delete_data_disks_on_termination",
// 					"delete_os_disk_on_termination",
// 				},
// 			},
// 		},
// 	})
// }
//
// func TestAccAzureStackVirtualMachine_importBasic_managedDisk(t *testing.T) {
// 	resourceName := "azurestack_virtual_machine.test"
//
// 	ri := acctest.RandInt()
// 	config := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_explicit(ri, testLocation())
//
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: config,
// 			},
//
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 				ImportStateVerifyIgnore: []string{
// 					"delete_data_disks_on_termination",
// 					"delete_os_disk_on_termination",
// 				},
// 			},
// 		},
// 	})
// }
