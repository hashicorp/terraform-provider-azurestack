package web

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2020-06-01/web"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/locks"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/web/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

var appServiceSourceControlTokenResourceName = "azurerm_app_service_source_control_token"

func resourceAppServiceSourceControlToken() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceAppServiceSourceControlTokenCreateUpdate,
		Read:   resourceAppServiceSourceControlTokenRead,
		Update: resourceAppServiceSourceControlTokenCreateUpdate,
		Delete: resourceAppServiceSourceControlTokenDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := validate.SourceControlTokenName()(id, "id")
			if len(err) > 0 {
				return fmt.Errorf("%s", err)
			}

			return nil
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"type": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.SourceControlTokenName(),
			},

			"token": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"token_secret": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func resourceAppServiceSourceControlTokenCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Web.BaseClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for App Service Source Control Token creation.")

	scmType := d.Get("type").(string)
	token := d.Get("token").(string)
	tokenSecret := d.Get("token_secret").(string)

	locks.ByName(scmType, appServiceSourceControlTokenResourceName)
	defer locks.UnlockByName(scmType, appServiceSourceControlTokenResourceName)

	properties := web.SourceControl{
		SourceControlProperties: &web.SourceControlProperties{
			Token:       utils.String(token),
			TokenSecret: utils.String(tokenSecret),
		},
	}

	if _, err := client.UpdateSourceControl(ctx, scmType, properties); err != nil {
		return fmt.Errorf("Error updating App Service Source Control Token (Type %q): %s", scmType, err)
	}

	read, err := client.GetSourceControl(ctx, scmType)
	if err != nil {
		return fmt.Errorf("Error retrieving App Service Source Control Token (Type %q): %s", scmType, err)
	}
	if read.Name == nil {
		return fmt.Errorf("Cannot read App Service Source Control Token (Type %q)", scmType)
	}

	d.SetId(*read.Name)

	return resourceAppServiceSourceControlTokenRead(d, meta)
}

func resourceAppServiceSourceControlTokenRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Web.BaseClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()
	scmType := d.Id()

	resp, err := client.GetSourceControl(ctx, scmType)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] App Service Source Control Token (Type %q) was not found - removing from state", scmType)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on App Service Source Control Token (Type %q): %+v", scmType, err)
	}

	d.Set("type", resp.Name)

	if props := resp.SourceControlProperties; props != nil {
		d.Set("token", props.Token)
		d.Set("token_secret", props.TokenSecret)
	}

	return nil
}

func resourceAppServiceSourceControlTokenDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Web.BaseClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	scmType := d.Id()

	// Delete cleans up existing tokens by setting their values to ""
	token := ""
	tokenSecret := ""

	locks.ByName(scmType, appServiceSourceControlTokenResourceName)
	defer locks.UnlockByName(scmType, appServiceSourceControlTokenResourceName)

	log.Printf("[DEBUG] Deleting App Service Source Control Token (Type %q)", scmType)

	properties := web.SourceControl{
		SourceControlProperties: &web.SourceControlProperties{
			Token:       utils.String(token),
			TokenSecret: utils.String(tokenSecret),
		},
	}

	if _, err := client.UpdateSourceControl(ctx, scmType, properties); err != nil {
		return fmt.Errorf("Error deleting App Service Source Control Token (Type %q): %s", scmType, err)
	}

	return nil
}
