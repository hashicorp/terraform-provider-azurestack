package datashare

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/datashare/mgmt/2019-11-01/datashare"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/datashare/helper"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/datashare/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/datashare/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func dataSourceDataShareDatasetDataLakeGen1() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceDataShareDatasetDataLakeGen1Read,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.DataSetName(),
			},

			"data_share_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.ShareID,
			},

			"data_lake_store_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"folder_path": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"file_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"display_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDataShareDatasetDataLakeGen1Read(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataShare.DataSetClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	shareID := d.Get("data_share_id").(string)
	shareId, err := parse.ShareID(shareID)
	if err != nil {
		return err
	}

	respModel, err := client.Get(ctx, shareId.ResourceGroup, shareId.AccountName, shareId.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving DataShare Data Lake Gen1 DataSet %q (Resource Group %q / accountName %q / shareName %q): %+v", name, shareId.ResourceGroup, shareId.AccountName, shareId.Name, err)
	}

	respId := helper.GetAzurermDataShareDataSetId(respModel.Value)
	if respId == nil || *respId == "" {
		return fmt.Errorf("empty or nil ID returned for DataShare Data Lake Gen1 DataSet %q (Resource Group %q / accountName %q / shareName %q)", name, shareId.ResourceGroup, shareId.AccountName, shareId.Name)
	}

	d.SetId(*respId)
	d.Set("name", name)
	d.Set("data_share_id", shareID)

	switch resp := respModel.Value.(type) {
	case datashare.ADLSGen1FileDataSet:
		if props := resp.ADLSGen1FileProperties; props != nil {
			if props.SubscriptionID != nil && props.ResourceGroup != nil && props.AccountName != nil {
				d.Set("data_lake_store_id", fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DataLakeStore/accounts/%s", *props.SubscriptionID, *props.ResourceGroup, *props.AccountName))
			}
			d.Set("folder_path", props.FolderPath)
			d.Set("file_name", props.FileName)
			d.Set("display_name", props.DataSetID)
		}

	case datashare.ADLSGen1FolderDataSet:
		if props := resp.ADLSGen1FolderProperties; props != nil {
			if props.SubscriptionID != nil && props.ResourceGroup != nil && props.AccountName != nil {
				d.Set("data_lake_store_id", fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DataLakeStore/accounts/%s", *props.SubscriptionID, *props.ResourceGroup, *props.AccountName))
			}
			d.Set("folder_path", props.FolderPath)
			d.Set("display_name", props.DataSetID)
		}

	default:
		return fmt.Errorf("data share dataset %q (Resource Group %q / accountName %q / shareName %q) is not a datalake store gen1 dataset", name, shareId.ResourceGroup, shareId.AccountName, shareId.Name)
	}

	return nil
}
