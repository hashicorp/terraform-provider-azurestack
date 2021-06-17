package kusto

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/kusto/mgmt/2020-09-18/kusto"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/kusto/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/kusto/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceKustoAttachedDatabaseConfiguration() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceKustoAttachedDatabaseConfigurationCreateUpdate,
		Read:   resourceKustoAttachedDatabaseConfigurationRead,
		Update: resourceKustoAttachedDatabaseConfigurationCreateUpdate,
		Delete: resourceKustoAttachedDatabaseConfigurationDelete,

		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.DataConnectionName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"cluster_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ClusterName,
			},

			"database_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.Any(validate.DatabaseName, validation.StringInSlice([]string{"*"}, false)),
			},

			"cluster_resource_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"attached_database_names": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"default_principal_modification_kind": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  kusto.DefaultPrincipalsModificationKindNone,
				ValidateFunc: validation.StringInSlice([]string{
					string(kusto.DefaultPrincipalsModificationKindNone),
					string(kusto.DefaultPrincipalsModificationKindReplace),
					string(kusto.DefaultPrincipalsModificationKindUnion),
				}, false),
			},
		},
	}
}

func resourceKustoAttachedDatabaseConfigurationCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.AttachedDatabaseConfigurationsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure Kusto Attached Database Configuration creation.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	clusterName := d.Get("cluster_name").(string)

	if d.IsNewResource() {
		resp, err := client.Get(ctx, resourceGroup, clusterName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Error checking for presence of existing Kusto Attached Database Configuration %q (Resource Group %q, Cluster %q): %s", name, resourceGroup, clusterName, err)
			}
		}

		if resp.ID != nil && *resp.ID != "" {
			return tf.ImportAsExistsError("azurerm_kusto_attached_database_configuration", *resp.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))

	configurationProperties := expandKustoAttachedDatabaseConfigurationProperties(d)

	configurationRequest := kusto.AttachedDatabaseConfiguration{
		Location:                                &location,
		AttachedDatabaseConfigurationProperties: configurationProperties,
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, clusterName, name, configurationRequest)
	if err != nil {
		return fmt.Errorf("Error creating or updating Kusto Attached Database Configuration %q (Resource Group %q, Cluster %q): %+v", name, resourceGroup, clusterName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for completion of Kusto Attached Database Configuration %q (Resource Group %q, Cluster %q): %+v", name, resourceGroup, clusterName, err)
	}

	configuration, err := client.Get(ctx, resourceGroup, clusterName, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Kusto Attached Database Configuration %q (Resource Group %q, Cluster %q): %+v", name, resourceGroup, clusterName, err)
	}

	d.SetId(*configuration.ID)

	return resourceKustoAttachedDatabaseConfigurationRead(d, meta)
}

func resourceKustoAttachedDatabaseConfigurationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.AttachedDatabaseConfigurationsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AttachedDatabaseConfigurationID(d.Id())
	if err != nil {
		return err
	}

	configuration, err := client.Get(ctx, id.ResourceGroup, id.ClusterName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(configuration.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Kusto Attached Database Configuration %q (Resource Group %q, Cluster %q): %+v", id.Name, id.ResourceGroup, id.ClusterName, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("cluster_name", id.ClusterName)

	if location := configuration.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := configuration.AttachedDatabaseConfigurationProperties; props != nil {
		d.Set("cluster_resource_id", props.ClusterResourceID)
		d.Set("database_name", props.DatabaseName)
		d.Set("default_principal_modification_kind", props.DefaultPrincipalsModificationKind)
		d.Set("attached_database_names", props.AttachedDatabaseNames)
	}

	return nil
}

func resourceKustoAttachedDatabaseConfigurationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.AttachedDatabaseConfigurationsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.AttachedDatabaseConfigurationID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.ClusterName, id.Name)
	if err != nil {
		return fmt.Errorf("Error deleting Kusto Attached Database Configuration %q (Resource Group %q, Cluster %q): %+v", id.Name, id.ResourceGroup, id.ClusterName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for deletion of Kusto Attached Database Configuration %q (Resource Group %q, Cluster %q): %+v", id.Name, id.ResourceGroup, id.ClusterName, err)
	}

	return nil
}

func expandKustoAttachedDatabaseConfigurationProperties(d *pluginsdk.ResourceData) *kusto.AttachedDatabaseConfigurationProperties {
	AttachedDatabaseConfigurationProperties := &kusto.AttachedDatabaseConfigurationProperties{}

	if clusterResourceID, ok := d.GetOk("cluster_resource_id"); ok {
		AttachedDatabaseConfigurationProperties.ClusterResourceID = utils.String(clusterResourceID.(string))
	}

	if databaseName, ok := d.GetOk("database_name"); ok {
		AttachedDatabaseConfigurationProperties.DatabaseName = utils.String(databaseName.(string))
	}

	if defaultPrincipalModificationKind, ok := d.GetOk("default_principal_modification_kind"); ok {
		AttachedDatabaseConfigurationProperties.DefaultPrincipalsModificationKind = kusto.DefaultPrincipalsModificationKind(defaultPrincipalModificationKind.(string))
	}

	return AttachedDatabaseConfigurationProperties
}
