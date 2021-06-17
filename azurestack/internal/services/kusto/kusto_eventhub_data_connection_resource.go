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
	eventhubValidate "github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/eventhub/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/kusto/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/kusto/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceKustoEventHubDataConnection() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceKustoEventHubDataConnectionCreateUpdate,
		Read:   resourceKustoEventHubDataConnectionRead,
		Update: resourceKustoEventHubDataConnectionCreateUpdate,
		Delete: resourceKustoEventHubDataConnectionDelete,

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.DataConnectionID(id)
			return err
		}, importDataConnection(kusto.KindEventHub)),

		Timeouts: &pluginsdk.ResourceTimeout{
			// TODO: confirm these
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

			"compression": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  kusto.CompressionNone,
				ValidateFunc: validation.StringInSlice([]string{
					string(kusto.CompressionGZip),
					string(kusto.CompressionNone),
				}, false),
			},

			"database_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.DatabaseName,
			},

			"eventhub_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: eventhubValidate.EventHubID,
			},

			"event_system_properties": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},

			"consumer_group": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: eventhubValidate.ValidateEventHubConsumerName(),
			},

			"table_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.EntityName,
			},

			"mapping_rule_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.EntityName,
			},

			"data_format": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(kusto.AVRO),
					string(kusto.CSV),
					string(kusto.JSON),
					string(kusto.MULTIJSON),
					string(kusto.PSV),
					string(kusto.RAW),
					string(kusto.SCSV),
					string(kusto.SINGLEJSON),
					string(kusto.SOHSV),
					string(kusto.TSV),
					string(kusto.TXT),
				}, false),
			},
		},
	}
}

func resourceKustoEventHubDataConnectionCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.DataConnectionsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure Kusto Event Hub Data Connection creation.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	clusterName := d.Get("cluster_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.IsNewResource() {
		connectionModel, err := client.Get(ctx, resourceGroup, clusterName, databaseName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(connectionModel.Response) {
				return fmt.Errorf("Error checking for presence of existing Kusto Event Hub Data Connection %q (Resource Group %q, Cluster %q, Database %q): %s", name, resourceGroup, clusterName, databaseName, err)
			}
		}

		if dataConnection, ok := connectionModel.Value.(kusto.EventHubDataConnection); ok {
			if dataConnection.ID != nil && *dataConnection.ID != "" {
				return tf.ImportAsExistsError("azurerm_kusto_eventhub_data_connection", *dataConnection.ID)
			}
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))

	eventHubDataConnectionProperties := expandKustoEventHubDataConnectionProperties(d)

	dataConnection1 := kusto.EventHubDataConnection{
		Name:                         &name,
		Location:                     &location,
		EventHubConnectionProperties: eventHubDataConnectionProperties,
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, clusterName, databaseName, name, dataConnection1)
	if err != nil {
		return fmt.Errorf("Error creating or updating Kusto Event Hub Data Connection %q (Resource Group %q, Cluster %q, Database: %q): %+v", name, resourceGroup, clusterName, databaseName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for completion of Kusto Event Hub Data Connection %q (Resource Group %q, Cluster %q, Database: %q): %+v", name, resourceGroup, clusterName, databaseName, err)
	}

	connectionModel, getDetailsErr := client.Get(ctx, resourceGroup, clusterName, databaseName, name)

	if getDetailsErr != nil {
		return fmt.Errorf("Error retrieving Kusto Event Hub Data Connection %q (Resource Group %q, Cluster %q, Database: %q): %+v", name, resourceGroup, clusterName, databaseName, err)
	}

	if dataConnection, ok := connectionModel.Value.(kusto.EventHubDataConnection); ok {
		if dataConnection.ID == nil {
			return fmt.Errorf("Cannot read ID for Kusto Event Hub Data Connection %q (Resource Group %q, Cluster %q, Database: %q): %+v", name, resourceGroup, clusterName, databaseName, err)
		}

		d.SetId(*dataConnection.ID)
	}

	return resourceKustoEventHubDataConnectionRead(d, meta)
}

func resourceKustoEventHubDataConnectionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.DataConnectionsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DataConnectionID(d.Id())
	if err != nil {
		return err
	}

	connectionModel, err := client.Get(ctx, id.ResourceGroup, id.ClusterName, id.DatabaseName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(connectionModel.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Kusto Event Hub Data Connection %q (Resource Group %q, Cluster %q, Database %q): %+v", id.Name, id.ResourceGroup, id.ClusterName, id.DatabaseName, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("cluster_name", id.ClusterName)
	d.Set("database_name", id.DatabaseName)

	if dataConnection, ok := connectionModel.Value.(kusto.EventHubDataConnection); ok {
		if location := dataConnection.Location; location != nil {
			d.Set("location", azure.NormalizeLocation(*location))
		}

		if props := dataConnection.EventHubConnectionProperties; props != nil {
			d.Set("eventhub_id", props.EventHubResourceID)
			d.Set("consumer_group", props.ConsumerGroup)
			d.Set("table_name", props.TableName)
			d.Set("mapping_rule_name", props.MappingRuleName)
			d.Set("data_format", props.DataFormat)
			d.Set("compression", props.Compression)
			d.Set("event_system_properties", props.EventSystemProperties)
		}
	}

	return nil
}

func resourceKustoEventHubDataConnectionDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Kusto.DataConnectionsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DataConnectionID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.ClusterName, id.DatabaseName, id.Name)
	if err != nil {
		return fmt.Errorf("Error deleting Kusto Event Hub Data Connection %q (Resource Group %q, Cluster %q, Database %q): %+v", id.Name, id.ResourceGroup, id.ClusterName, id.DatabaseName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for deletion of Kusto Event Hub Data Connection %q (Resource Group %q, Cluster %q, Database %q): %+v", id.Name, id.ResourceGroup, id.ClusterName, id.DatabaseName, err)
	}

	return nil
}

func expandKustoEventHubDataConnectionProperties(d *pluginsdk.ResourceData) *kusto.EventHubConnectionProperties {
	eventHubConnectionProperties := &kusto.EventHubConnectionProperties{}

	if eventhubResourceID, ok := d.GetOk("eventhub_id"); ok {
		eventHubConnectionProperties.EventHubResourceID = utils.String(eventhubResourceID.(string))
	}

	if consumerGroup, ok := d.GetOk("consumer_group"); ok {
		eventHubConnectionProperties.ConsumerGroup = utils.String(consumerGroup.(string))
	}

	if tableName, ok := d.GetOk("table_name"); ok {
		eventHubConnectionProperties.TableName = utils.String(tableName.(string))
	}

	if mappingRuleName, ok := d.GetOk("mapping_rule_name"); ok {
		eventHubConnectionProperties.MappingRuleName = utils.String(mappingRuleName.(string))
	}

	if df, ok := d.GetOk("data_format"); ok {
		eventHubConnectionProperties.DataFormat = kusto.EventHubDataFormat(df.(string))
	}

	if compression, ok := d.GetOk("compression"); ok {
		eventHubConnectionProperties.Compression = kusto.Compression(compression.(string))
	}

	if eventSystemProperties, ok := d.GetOk("event_system_properties"); ok {
		props := make([]string, 0)
		for _, prop := range eventSystemProperties.([]interface{}) {
			props = append(props, prop.(string))
		}
		eventHubConnectionProperties.EventSystemProperties = &props
	}

	return eventHubConnectionProperties
}
