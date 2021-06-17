package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/botservice/mgmt/2018-07-12/botservice"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/location"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/bot/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceBotChannelsRegistration() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceBotChannelsRegistrationCreate,
		Read:   resourceBotChannelsRegistrationRead,
		Delete: resourceBotChannelsRegistrationDelete,
		Update: resourceBotChannelsRegistrationUpdate,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.BotServiceID(id)
			return err
		}, func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) ([]*pluginsdk.ResourceData, error) {
			client := meta.(*clients.Client).Bot.BotClient

			id, err := parse.BotServiceID(d.Id())
			if err != nil {
				return nil, err
			}

			resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
			if err != nil {
				if utils.ResponseWasNotFound(resp.Response) {
					return nil, fmt.Errorf("Bot Channels Registration %q was not found in Resource Group %q", id.Name, id.ResourceGroup)
				}

				return nil, fmt.Errorf("retrieving Bot Channels Registration %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
			}
			if resp.Kind != botservice.KindBot {
				return nil, fmt.Errorf("Bot %q (Resource Group %q) was not a Channel Registration - got %q", id.Name, id.ResourceGroup, string(resp.Kind))
			}

			return []*pluginsdk.ResourceData{d}, nil
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"sku": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(botservice.F0),
					string(botservice.S1),
				}, false),
			},

			"microsoft_app_id": {
				Type:         pluginsdk.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},

			"display_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"endpoint": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"developer_app_insights_key": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsUUID,
			},

			"developer_app_insights_api_key": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"developer_app_insights_application_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsUUID,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceBotChannelsRegistrationCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Bot.BotClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceId := parse.NewBotServiceID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Bot Channels Registration %q (Resource Group %q): %+v", resourceId.Name, resourceId.ResourceGroup, err)
			}
		}
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_bot_channels_registration", resourceId.ID())
		}
	}

	displayName := d.Get("display_name").(string)
	if displayName == "" {
		displayName = resourceId.Name
	}

	bot := botservice.Bot{
		Properties: &botservice.BotProperties{
			DisplayName:                       utils.String(displayName),
			Endpoint:                          utils.String(d.Get("endpoint").(string)),
			MsaAppID:                          utils.String(d.Get("microsoft_app_id").(string)),
			DeveloperAppInsightKey:            utils.String(d.Get("developer_app_insights_key").(string)),
			DeveloperAppInsightsAPIKey:        utils.String(d.Get("developer_app_insights_api_key").(string)),
			DeveloperAppInsightsApplicationID: utils.String(d.Get("developer_app_insights_application_id").(string)),
		},
		Location: utils.String(d.Get("location").(string)),
		Sku: &botservice.Sku{
			Name: botservice.SkuName(d.Get("sku").(string)),
		},
		Kind: botservice.KindBot,
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if _, err := client.Create(ctx, resourceId.ResourceGroup, resourceId.Name, bot); err != nil {
		return fmt.Errorf("creating Bot Channels Registration %q (Resource Group %q): %+v", resourceId.Name, resourceId.ResourceGroup, err)
	}

	d.SetId(resourceId.ID())
	return resourceBotChannelsRegistrationRead(d, meta)
}

func resourceBotChannelsRegistrationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Bot.BotClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BotServiceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Bot Channels Registration %q (Resource Group %q) was not found - removing from state", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Bot Channels Registration %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if sku := resp.Sku; sku != nil {
		d.Set("sku", string(sku.Name))
	}

	if props := resp.Properties; props != nil {
		d.Set("microsoft_app_id", props.MsaAppID)
		d.Set("endpoint", props.Endpoint)
		d.Set("display_name", props.DisplayName)
		d.Set("developer_app_insights_key", props.DeveloperAppInsightKey)
		d.Set("developer_app_insights_application_id", props.DeveloperAppInsightsApplicationID)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceBotChannelsRegistrationUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Bot.BotClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BotServiceID(d.Id())
	if err != nil {
		return err
	}

	t := d.Get("tags").(map[string]interface{})
	displayName := d.Get("display_name").(string)
	if displayName == "" {
		displayName = id.Name
	}

	bot := botservice.Bot{
		Properties: &botservice.BotProperties{
			DisplayName:                       utils.String(displayName),
			Endpoint:                          utils.String(d.Get("endpoint").(string)),
			MsaAppID:                          utils.String(d.Get("microsoft_app_id").(string)),
			DeveloperAppInsightKey:            utils.String(d.Get("developer_app_insights_key").(string)),
			DeveloperAppInsightsAPIKey:        utils.String(d.Get("developer_app_insights_api_key").(string)),
			DeveloperAppInsightsApplicationID: utils.String(d.Get("developer_app_insights_application_id").(string)),
		},
		Location: utils.String(d.Get("location").(string)),
		Sku: &botservice.Sku{
			Name: botservice.SkuName(d.Get("sku").(string)),
		},
		Kind: botservice.KindBot,
		Tags: tags.Expand(t),
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.Name, bot); err != nil {
		return fmt.Errorf("updating Bot Channels Registration %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return resourceBotChannelsRegistrationRead(d, meta)
}

func resourceBotChannelsRegistrationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Bot.BotClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.BotServiceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting Bot Channels Registration %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	return nil
}
