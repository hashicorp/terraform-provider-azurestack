package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-11-01/network"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/locks"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/network/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceSubnetNetworkSecurityGroupAssociation() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceSubnetNetworkSecurityGroupAssociationCreate,
		Read:   resourceSubnetNetworkSecurityGroupAssociationRead,
		Delete: resourceSubnetNetworkSecurityGroupAssociationDelete,
		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"subnet_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"network_security_group_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},
		},
	}
}

func resourceSubnetNetworkSecurityGroupAssociationCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SubnetsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Subnet <-> Network Security Group Association creation.")

	subnetId := d.Get("subnet_id").(string)
	networkSecurityGroupId := d.Get("network_security_group_id").(string)

	parsedSubnetId, err := azure.ParseAzureResourceID(subnetId)
	if err != nil {
		return err
	}

	parsedNetworkSecurityGroupId, err := parse.NetworkSecurityGroupID(networkSecurityGroupId)
	if err != nil {
		return err
	}

	locks.ByName(parsedNetworkSecurityGroupId.Name, networkSecurityGroupResourceName)
	defer locks.UnlockByName(parsedNetworkSecurityGroupId.Name, networkSecurityGroupResourceName)

	subnetName := parsedSubnetId.Path["subnets"]
	virtualNetworkName := parsedSubnetId.Path["virtualNetworks"]
	resourceGroup := parsedSubnetId.ResourceGroup

	locks.ByName(subnetName, SubnetResourceName)
	defer locks.UnlockByName(subnetName, SubnetResourceName)

	locks.ByName(virtualNetworkName, VirtualNetworkResourceName)
	defer locks.UnlockByName(virtualNetworkName, VirtualNetworkResourceName)

	subnet, err := client.Get(ctx, resourceGroup, virtualNetworkName, subnetName, "")
	if err != nil {
		if utils.ResponseWasNotFound(subnet.Response) {
			return fmt.Errorf("subnet %q (Virtual Network %q / Resource Group %q) was not found", subnetName, virtualNetworkName, resourceGroup)
		}

		return fmt.Errorf("retrieving Subnet %q (Virtual Network %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	if props := subnet.SubnetPropertiesFormat; props != nil {
		if nsg := props.NetworkSecurityGroup; nsg != nil {
			// we're intentionally not checking the ID - if there's a NSG, it needs to be imported
			if nsg.ID != nil && subnet.ID != nil {
				return tf.ImportAsExistsError("azurerm_subnet_network_security_group_association", *subnet.ID)
			}
		}

		props.NetworkSecurityGroup = &network.SecurityGroup{
			ID: utils.String(networkSecurityGroupId),
		}
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, virtualNetworkName, subnetName, subnet)
	if err != nil {
		return fmt.Errorf("updating Network Security Group Association for Subnet %q (Virtual Network %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of Network Security Group  Association for Subnet %q (VN %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, virtualNetworkName, subnetName, "")
	if err != nil {
		return fmt.Errorf("retrieving Subnet %q (Virtual Network %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	d.SetId(*read.ID)

	return resourceSubnetNetworkSecurityGroupAssociationRead(d, meta)
}

func resourceSubnetNetworkSecurityGroupAssociationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SubnetsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	virtualNetworkName := id.Path["virtualNetworks"]
	subnetName := id.Path["subnets"]

	resp, err := client.Get(ctx, resourceGroup, virtualNetworkName, subnetName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Subnet %q (Virtual Network %q / Resource Group %q) could not be found - removing from state!", subnetName, virtualNetworkName, resourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Subnet %q (Virtual Network %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	props := resp.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("`properties` was nil for Subnet %q (Virtual Network %q / Resource Group %q)", subnetName, virtualNetworkName, resourceGroup)
	}

	securityGroup := props.NetworkSecurityGroup
	if securityGroup == nil {
		log.Printf("[DEBUG] Subnet %q (Virtual Network %q / Resource Group %q) doesn't have a Network Security Group - removing from state!", subnetName, virtualNetworkName, resourceGroup)
		d.SetId("")
		return nil
	}

	d.Set("subnet_id", resp.ID)
	d.Set("network_security_group_id", securityGroup.ID)

	return nil
}

func resourceSubnetNetworkSecurityGroupAssociationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SubnetsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	virtualNetworkName := id.Path["virtualNetworks"]
	subnetName := id.Path["subnets"]

	// retrieve the subnet
	read, err := client.Get(ctx, resourceGroup, virtualNetworkName, subnetName, "")
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			log.Printf("[DEBUG] Subnet %q (Virtual Network %q / Resource Group %q) could not be found - removing from state!", subnetName, virtualNetworkName, resourceGroup)
			return nil
		}

		return fmt.Errorf("retrieving Subnet %q (Virtual Network %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	props := read.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("`Properties` was nil for Subnet %q (Virtual Network %q / Resource Group %q)", subnetName, virtualNetworkName, resourceGroup)
	}

	if props.NetworkSecurityGroup == nil || props.NetworkSecurityGroup.ID == nil {
		log.Printf("[DEBUG] Subnet %q (Virtual Network %q / Resource Group %q) has no Network Security Group - removing from state!", subnetName, virtualNetworkName, resourceGroup)
		return nil
	}

	// once we have the network security group id to lock on, lock on that
	parsedNetworkSecurityGroupId, err := parse.NetworkSecurityGroupID(*props.NetworkSecurityGroup.ID)
	if err != nil {
		return err
	}

	locks.ByName(parsedNetworkSecurityGroupId.Name, networkSecurityGroupResourceName)
	defer locks.UnlockByName(parsedNetworkSecurityGroupId.Name, networkSecurityGroupResourceName)

	locks.ByName(virtualNetworkName, VirtualNetworkResourceName)
	defer locks.UnlockByName(virtualNetworkName, VirtualNetworkResourceName)

	locks.ByName(subnetName, SubnetResourceName)
	defer locks.UnlockByName(subnetName, SubnetResourceName)

	// then re-retrieve it to ensure we've got the latest state
	read, err = client.Get(ctx, resourceGroup, virtualNetworkName, subnetName, "")
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			log.Printf("[DEBUG] Subnet %q (Virtual Network %q / Resource Group %q) could not be found - removing from state!", subnetName, virtualNetworkName, resourceGroup)
			return nil
		}

		return fmt.Errorf("retrieving Subnet %q (Virtual Network %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	read.SubnetPropertiesFormat.NetworkSecurityGroup = nil

	future, err := client.CreateOrUpdate(ctx, resourceGroup, virtualNetworkName, subnetName, read)
	if err != nil {
		return fmt.Errorf("removing Network Security Group Association from Subnet %q (Virtual Network %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for removal of Network Security Group Association from Subnet %q (Virtual Network %q / Resource Group %q): %+v", subnetName, virtualNetworkName, resourceGroup, err)
	}

	return nil
}
