package compute_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	networkParse "github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/ssh"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type ImageResource struct{}

func TestImageReource_standaloneImage(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_image", "test")
	r := ImageResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// need to create a vm and then reference it in the image creation
			Config: r.setupUnmanagedDisks(data, "LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				data.CheckWithClientForResource(r.virtualMachineExists, "azurestack_virtual_machine.testsource"),
			),
		},
		{
			Config: r.standaloneImageProvision(data, "LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r ImageResource) standaloneImageProvision(data acceptance.TestData, storageType string) string {
	template := r.setupUnmanagedDisks(data, storageType)
	return fmt.Sprintf(`
%s

resource "azurestack_image" "test" {
  name                = "accteste"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  zone_resilient      = false

  os_disk {
    os_type  = "Linux"
    os_state = "Generalized"
    blob_uri = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/myosdisk1.vhd"
    size_gb  = 30
    caching  = "None"
  }
}
`, template)
}

func (ImageResource) generalizeVirtualMachine(data acceptance.TestData) func(context.Context, *clients.Client, *pluginsdk.InstanceState) error {
	return func(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
		id, err := parse.VirtualMachineID(state.ID)
		if err != nil {
			return err
		}

		// these are nested in a Set in the Legacy VM resource, simpler to compute them
		userName := fmt.Sprintf("testadmin%d", data.RandomInteger)
		password := fmt.Sprintf("Password1234!%d", data.RandomInteger)

		// first retrieve the Virtual Machine, since we need to find
		nicIdRaw := state.Attributes["network_interface_ids.0"]
		nicId, err := networkParse.NetworkInterfaceID(nicIdRaw)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Retrieving Network Interface..")
		nic, err := client.Network.InterfacesClient.Get(ctx, nicId.ResourceGroup, nicId.Name, "")
		if err != nil {
			return fmt.Errorf("retrieving %s: %+v", *nicId, err)
		}

		publicIpRaw := ""
		if props := nic.InterfacePropertiesFormat; props != nil {
			if configs := props.IPConfigurations; configs != nil {
				for _, config := range *props.IPConfigurations {
					if config.InterfaceIPConfigurationPropertiesFormat == nil {
						continue
					}

					if config.InterfaceIPConfigurationPropertiesFormat.PublicIPAddress == nil {
						continue
					}

					if config.InterfaceIPConfigurationPropertiesFormat.PublicIPAddress.ID == nil {
						continue
					}

					publicIpRaw = *config.InterfaceIPConfigurationPropertiesFormat.PublicIPAddress.ID
					break
				}
			}
		}
		if publicIpRaw == "" {
			return fmt.Errorf("retrieving %s: could not determine Public IP Address ID", *nicId)
		}

		log.Printf("[DEBUG] Retrieving Public IP Address %q..", publicIpRaw)
		publicIpId, err := networkParse.PublicIpAddressID(publicIpRaw)
		if err != nil {
			return err
		}

		publicIpAddress, err := client.Network.PublicIPsClient.Get(ctx, publicIpId.ResourceGroup, publicIpId.Name, "")
		if err != nil {
			return fmt.Errorf("retrieving %s: %+v", *publicIpId, err)
		}
		fqdn := ""
		if props := publicIpAddress.PublicIPAddressPropertiesFormat; props != nil {
			if dns := props.DNSSettings; dns != nil {
				if dns.Fqdn != nil {
					fqdn = *dns.Fqdn
				}
			}
		}
		if fqdn == "" {
			return fmt.Errorf("unable to determine FQDN for %q", *publicIpId)
		}

		log.Printf("[DEBUG] Running Generalization Command..")
		sshGeneralizationCommand := ssh.Runner{
			Hostname: fqdn,
			Port:     22,
			Username: userName,
			Password: password,
			CommandsToRun: []string{
				ssh.LinuxAgentDeprovisionCommand,
			},
		}
		if err := sshGeneralizationCommand.Run(ctx); err != nil {
			return fmt.Errorf("Bad: running generalization command: %+v", err)
		}

		log.Printf("[DEBUG] Deallocating VM..")
		// Upgrading to the 2021-07-01 exposed a new hibernate parameter in the GET method
		future, err := client.Compute.VMClient.Deallocate(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			return fmt.Errorf("Bad: deallocating vm: %+v", err)
		}
		log.Printf("[DEBUG] Waiting for Deallocation..")
		if err = future.WaitForCompletionRef(ctx, client.Compute.VMClient.Client); err != nil {
			return fmt.Errorf("Bad: waiting for deallocation: %+v", err)
		}

		log.Printf("[DEBUG] Generalizing VM..")
		if _, err = client.Compute.VMClient.Generalize(ctx, id.ResourceGroup, id.Name); err != nil {
			return fmt.Errorf("Bad: Generalizing error %+v", err)
		}

		return nil
	}
}

func (ImageResource) virtualMachineExists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	id, err := parse.VirtualMachineID(state.ID)
	if err != nil {
		return err
	}

	resp, err := client.Compute.VMClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s does not exist", *id)
		}

		return fmt.Errorf("Bad: Get on client: %+v", err)
	}

	return nil
}

func (ImageResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ImageID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.ImageClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Image %q", id)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (r ImageResource) setupUnmanagedDisks(data acceptance.TestData, storageType string) string {
	template := r.template(data)
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}
%s
resource "azurestack_network_interface" "testsource" {
  name                = "acctnicsource-${local.number}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  ip_configuration {
    name                          = "testconfigurationsource"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurestack_public_ip.test.id
  }
}
resource "azurestack_storage_account" "test" {
  name                     = "accsa${local.random_string}"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "%s"
}
resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "blob"
}

resource "azurestack_virtual_machine" "testsource" {
  name                  = "testsource"
  location              = azurestack_resource_group.test.location
  resource_group_name   = azurestack_resource_group.test.name
  network_interface_ids = [azurestack_network_interface.testsource.id]
  vm_size               = "Standard_F2s_v2"
  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
  storage_os_disk {
    name          = "myosdisk1"
    vhd_uri       = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/myosdisk1.vhd"
    caching       = "ReadWrite"
    create_option = "FromImage"
  }
  os_profile {
    computer_name  = "mdimagetestsource"
    admin_username = local.admin_username
    admin_password = local.admin_password
  }
  os_profile_linux_config {
    disable_password_authentication = false
  }
}
`, template, storageType)
}

func (ImageResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
locals {
  number         = "%d"
  location       = %q
  random_string  = %q
  admin_username = "testadmin%d"
  admin_password = "Password1234!%d"
}
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-${local.number}"
  location = local.location
}
resource "azurestack_virtual_network" "test" {
  name                = "acctvn-${local.number}"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  address_space       = ["10.0.0.0/16"]
}
resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}
resource "azurestack_public_ip" "test" {
  name                = "acctpip-${local.number}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, data.RandomInteger)
}
