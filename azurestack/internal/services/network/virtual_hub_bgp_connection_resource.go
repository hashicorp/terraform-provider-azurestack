package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-11-01/network"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/locks"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/network/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/network/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceVirtualHubBgpConnection() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceVirtualHubBgpConnectionCreate,
		Read:   resourceVirtualHubBgpConnectionRead,
		Delete: resourceVirtualHubBgpConnectionDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.BgpConnectionID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"virtual_hub_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.VirtualHubID,
			},

			"peer_asn": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"peer_ip": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
		},
	}
}

func resourceVirtualHubBgpConnectionCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VirtualHubBgpConnectionClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualHubID(d.Get("virtual_hub_id").(string))
	if err != nil {
		return err
	}

	locks.ByName(id.Name, virtualHubResourceName)
	defer locks.UnlockByName(id.Name, virtualHubResourceName)

	name := d.Get("name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for present of existing Virtual Hub Bgp Connection %q (Resource Group %q / Virtual Hub %q): %+v", name, id.ResourceGroup, id.Name, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_virtual_hub_bgp_connection", *existing.ID)
		}
	}

	parameters := network.BgpConnection{
		Name: utils.String(d.Get("name").(string)),
		BgpConnectionProperties: &network.BgpConnectionProperties{
			PeerAsn: utils.Int64(int64(d.Get("peer_asn").(int))),
			PeerIP:  utils.String(d.Get("peer_ip").(string)),
		},
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating Virtual Hub Bgp Connection %q (Resource Group %q / Virtual Hub %q): %+v", name, id.ResourceGroup, id.Name, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on creating/updating future for Virtual Hub Bgp Connection %q (Resource Group %q / Virtual Hub %q): %+v", name, id.ResourceGroup, id.Name, err)
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving Virtual Hub Bgp Connection %q (Resource Group %q / Virtual Hub %q): %+v", name, id.ResourceGroup, id.Name, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Virtual Hub Bgp Connection %q (Resource Group %q / Virtual Hub %q) ID", name, id.ResourceGroup, id.Name)
	}

	d.SetId(*resp.ID)

	return resourceVirtualHubBgpConnectionRead(d, meta)
}

func resourceVirtualHubBgpConnectionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VirtualHubBgpConnectionClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BgpConnectionID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] network %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Virtual Hub Bgp Connection %q (Resource Group %q / Virtual Hub %q): %+v", id.Name, id.ResourceGroup, id.VirtualHubName, err)
	}

	d.Set("name", id.Name)
	d.Set("virtual_hub_id", parse.NewVirtualHubID(id.SubscriptionId, id.ResourceGroup, id.VirtualHubName).ID())

	if props := resp.BgpConnectionProperties; props != nil {
		d.Set("peer_asn", props.PeerAsn)
		d.Set("peer_ip", props.PeerIP)
	}

	return nil
}

func resourceVirtualHubBgpConnectionDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VirtualHubBgpConnectionClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BgpConnectionID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.VirtualHubName, virtualHubResourceName)
	defer locks.UnlockByName(id.VirtualHubName, virtualHubResourceName)

	future, err := client.Delete(ctx, id.ResourceGroup, id.VirtualHubName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Virtual Hub Bgp Connection %q (Resource Group %q / Virtual Hub %q): %+v", id.Name, id.ResourceGroup, id.VirtualHubName, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on deleting future for Virtual Hub Bgp Connection %q (Resource Group %q / Virtual Hub %q): %+v", id.Name, id.ResourceGroup, id.VirtualHubName, err)
	}

	return nil
}
