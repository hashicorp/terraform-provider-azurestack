package automation

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/automation/mgmt/2018-06-30-preview/automation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/automation/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceAutomationCredential() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceAutomationCredentialCreateUpdate,
		Read:   resourceAutomationCredentialRead,
		Update: resourceAutomationCredentialCreateUpdate,
		Delete: resourceAutomationCredentialDelete,

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
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"automation_account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AutomationAccount(),
			},

			"username": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"password": {
				Type:      pluginsdk.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAutomationCredentialCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Automation.CredentialClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM Automation Credential creation.")

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)
	accountName := d.Get("automation_account_name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resGroup, accountName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Automation Credential %q (Account %q / Resource Group %q): %s", name, accountName, resGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_automation_credential", *existing.ID)
		}
	}

	user := d.Get("username").(string)
	password := d.Get("password").(string)
	description := d.Get("description").(string)

	parameters := automation.CredentialCreateOrUpdateParameters{
		CredentialCreateOrUpdateProperties: &automation.CredentialCreateOrUpdateProperties{
			UserName:    &user,
			Password:    &password,
			Description: &description,
		},
		Name: &name,
	}

	if _, err := client.CreateOrUpdate(ctx, resGroup, accountName, name, parameters); err != nil {
		return err
	}

	read, err := client.Get(ctx, resGroup, accountName, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Automation Credential '%s' (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceAutomationCredentialRead(d, meta)
}

func resourceAutomationCredentialRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Automation.CredentialClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	accountName := id.Path["automationAccounts"]
	name := id.Path["credentials"]

	resp, err := client.Get(ctx, resGroup, accountName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on AzureRM Automation Credential '%s': %+v", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	d.Set("automation_account_name", accountName)
	if props := resp.CredentialProperties; props != nil {
		d.Set("username", props.UserName)
	}
	d.Set("description", resp.Description)

	return nil
}

func resourceAutomationCredentialDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Automation.CredentialClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	accountName := id.Path["automationAccounts"]
	name := id.Path["credentials"]

	resp, err := client.Delete(ctx, resGroup, accountName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("Error issuing AzureRM delete request for Automation Credential '%s': %+v", name, err)
	}

	return nil
}
