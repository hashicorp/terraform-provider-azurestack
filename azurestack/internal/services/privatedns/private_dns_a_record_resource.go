package privatedns

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/privatedns/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourcePrivateDnsARecord() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourcePrivateDnsARecordCreateUpdate,
		Read:   resourcePrivateDnsARecordRead,
		Update: resourcePrivateDnsARecordCreateUpdate,
		Delete: resourcePrivateDnsARecordDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ARecordID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				// lower-cased due to the broken API https://github.com/Azure/azure-rest-api-specs/issues/6641
				ValidateFunc: validate.LowerCasedString,
			},

			// TODO: make this case sensitive once the API's fixed https://github.com/Azure/azure-rest-api-specs/issues/6641
			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"zone_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"records": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:      pluginsdk.HashString,
			},

			"ttl": {
				Type:     pluginsdk.TypeInt,
				Required: true,
			},

			"fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourcePrivateDnsARecordCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).PrivateDns.RecordSetsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceId := parse.NewARecordID(subscriptionId, d.Get("resource_group_name").(string), d.Get("zone_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.PrivateDnsZoneName, privatedns.A, resourceId.AName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", resourceId, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_private_dns_a_record", *existing.ID)
		}
	}

	ttl := int64(d.Get("ttl").(int))
	t := d.Get("tags").(map[string]interface{})

	parameters := privatedns.RecordSet{
		Name: utils.String(resourceId.AName),
		RecordSetProperties: &privatedns.RecordSetProperties{
			Metadata: tags.Expand(t),
			TTL:      &ttl,
			ARecords: expandAzureRmPrivateDnsARecords(d),
		},
	}

	eTag := ""
	ifNoneMatch := "" // set to empty to allow updates to records after creation
	if _, err := client.CreateOrUpdate(ctx, resourceId.ResourceGroup, resourceId.PrivateDnsZoneName, privatedns.A, resourceId.AName, parameters, eTag, ifNoneMatch); err != nil {
		return fmt.Errorf("creating/updating %s: %+v", resourceId, err)
	}

	d.SetId(resourceId.ID())
	return resourcePrivateDnsARecordRead(d, meta)
}

func resourcePrivateDnsARecordRead(d *pluginsdk.ResourceData, meta interface{}) error {
	dnsClient := meta.(*clients.Client).PrivateDns.RecordSetsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ARecordID(d.Id())
	if err != nil {
		return err
	}

	resp, err := dnsClient.Get(ctx, id.ResourceGroup, id.PrivateDnsZoneName, privatedns.A, id.AName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.AName)
	d.Set("zone_name", id.PrivateDnsZoneName)
	d.Set("resource_group_name", id.ResourceGroup)

	d.Set("ttl", resp.TTL)
	d.Set("fqdn", resp.Fqdn)

	if err := d.Set("records", flattenAzureRmPrivateDnsARecords(resp.ARecords)); err != nil {
		return err
	}

	return tags.FlattenAndSet(d, resp.Metadata)
}

func resourcePrivateDnsARecordDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	dnsClient := meta.(*clients.Client).PrivateDns.RecordSetsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ARecordID(d.Id())
	if err != nil {
		return err
	}

	if _, err := dnsClient.Delete(ctx, id.ResourceGroup, id.PrivateDnsZoneName, privatedns.A, id.AName, ""); err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	return nil
}

func flattenAzureRmPrivateDnsARecords(records *[]privatedns.ARecord) []string {
	results := make([]string, 0)
	if records == nil {
		return results
	}

	for _, record := range *records {
		if record.Ipv4Address == nil {
			continue
		}

		results = append(results, *record.Ipv4Address)
	}

	return results
}

func expandAzureRmPrivateDnsARecords(d *pluginsdk.ResourceData) *[]privatedns.ARecord {
	recordStrings := d.Get("records").(*pluginsdk.Set).List()
	records := make([]privatedns.ARecord, len(recordStrings))

	for i, v := range recordStrings {
		ipv4 := v.(string)
		records[i] = privatedns.ARecord{
			Ipv4Address: &ipv4,
		}
	}

	return &records
}
