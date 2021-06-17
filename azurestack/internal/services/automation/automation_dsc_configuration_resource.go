package automation

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/automation/mgmt/2018-06-30-preview/automation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/automation/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceAutomationDscConfiguration() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceAutomationDscConfigurationCreateUpdate,
		Read:   resourceAutomationDscConfigurationRead,
		Update: resourceAutomationDscConfigurationCreateUpdate,
		Delete: resourceAutomationDscConfigurationDelete,

		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

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
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[a-zA-Z0-9_]{1,64}$`),
					`The name length must be from 1 to 64 characters. The name can only contain letters, numbers and underscores.`,
				),
			},

			"automation_account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AutomationAccount(),
			},

			"content_embedded": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"log_verbose": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"state": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceAutomationDscConfigurationCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Automation.DscConfigurationClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM Automation Dsc Configuration creation.")

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)
	accName := d.Get("automation_account_name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resGroup, accName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Automation DSC Configuration %q (Account %q / Resource Group %q): %s", name, accName, resGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_automation_dsc_configuration", *existing.ID)
		}
	}

	contentEmbedded := d.Get("content_embedded").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))
	logVerbose := d.Get("log_verbose").(bool)
	description := d.Get("description").(string)
	t := d.Get("tags").(map[string]interface{})

	parameters := automation.DscConfigurationCreateOrUpdateParameters{
		DscConfigurationCreateOrUpdateProperties: &automation.DscConfigurationCreateOrUpdateProperties{
			LogVerbose:  utils.Bool(logVerbose),
			Description: utils.String(description),
			Source: &automation.ContentSource{
				Type:  automation.EmbeddedContent,
				Value: utils.String(contentEmbedded),
			},
		},
		Location: utils.String(location),
		Tags:     tags.Expand(t),
	}

	if _, err := client.CreateOrUpdate(ctx, resGroup, accName, name, parameters); err != nil {
		return err
	}

	read, err := client.Get(ctx, resGroup, accName, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Automation Dsc Configuration %q (resource group %q) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceAutomationDscConfigurationRead(d, meta)
}

func resourceAutomationDscConfigurationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Automation.DscConfigurationClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	accName := id.Path["automationAccounts"]
	name := id.Path["configurations"]

	resp, err := client.Get(ctx, resGroup, accName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on AzureRM Automation Dsc Configuration %q: %+v", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	d.Set("automation_account_name", accName)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.DscConfigurationProperties; props != nil {
		d.Set("log_verbose", props.LogVerbose)
		d.Set("description", props.Description)
		d.Set("state", resp.State)
	}

	contentresp, err := client.GetContent(ctx, resGroup, accName, name)
	if err != nil {
		return fmt.Errorf("Error making Read request on AzureRM Automation Dsc Configuration content %q: %+v", name, err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(contentresp.Body); err != nil {
		return fmt.Errorf("Error reading from AzureRM Automation Dsc Configuration buffer %q: %+v", name, err)
	}
	content := buf.String()

	d.Set("content_embedded", content)

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceAutomationDscConfigurationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Automation.DscConfigurationClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	accName := id.Path["automationAccounts"]
	name := id.Path["configurations"]

	resp, err := client.Delete(ctx, resGroup, accName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("Error issuing AzureRM delete request for Automation Dsc Configuration %q: %+v", name, err)
	}

	return nil
}
