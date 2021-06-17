package springcloud

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/appplatform/mgmt/2020-11-01-preview/appplatform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/identity"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/springcloud/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/springcloud/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

type springCloudAppIdentity = identity.SystemAssigned

func resourceSpringCloudApp() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceSpringCloudAppCreate,
		Read:   resourceSpringCloudAppRead,
		Update: resourceSpringCloudAppUpdate,
		Delete: resourceSpringCloudAppDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SpringCloudAppID(id)
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
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SpringCloudAppName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"service_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SpringCloudServiceName,
			},

			"identity": springCloudAppIdentity{}.Schema(),

			"is_public": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"https_only": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"persistent_disk": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"size_in_gb": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 50),
						},

						"mount_path": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Default:      "/persistent",
							ValidateFunc: validate.MountPath,
						},
					},
				},
			},

			"tls_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"url": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSpringCloudAppCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.AppsClient
	servicesClient := meta.(*clients.Client).AppPlatform.ServicesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	serviceName := d.Get("service_name").(string)

	serviceResp, err := servicesClient.Get(ctx, resourceGroup, serviceName)
	if err != nil {
		return fmt.Errorf("unable to retrieve Spring Cloud Service %q (Resource Group %q): %+v", serviceName, resourceGroup, err)
	}

	resourceId := parse.NewSpringCloudAppID(subscriptionId, resourceGroup, serviceName, name).ID()
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, serviceName, name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Spring Cloud App %q (Spring Cloud Service %q / Resource Group %q): %+v", name, serviceName, resourceGroup, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_spring_cloud_app", resourceId)
		}
	}

	identity, err := expandSpringCloudAppIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return err
	}

	app := appplatform.AppResource{
		Location: serviceResp.Location,
		Identity: identity,
		Properties: &appplatform.AppResourceProperties{
			EnableEndToEndTLS: utils.Bool(d.Get("tls_enabled").(bool)),
			Public:            utils.Bool(d.Get("is_public").(bool)),
		},
	}
	future, err := client.CreateOrUpdate(ctx, resourceGroup, serviceName, name, app)
	if err != nil {
		return fmt.Errorf("creating Spring Cloud App %q (Spring Cloud Service %q / Resource Group %q): %+v", name, serviceName, resourceGroup, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of Spring Cloud App %q (Spring Cloud Service %q / Resource Group %q): %+v", name, serviceName, resourceGroup, err)
	}

	// HTTPSOnly and PersistentDisk could only be set by update
	app.Properties.HTTPSOnly = utils.Bool(d.Get("https_only").(bool))
	app.Properties.PersistentDisk = expandSpringCloudAppPersistentDisk(d.Get("persistent_disk").([]interface{}))
	future, err = client.CreateOrUpdate(ctx, resourceGroup, serviceName, name, app)
	if err != nil {
		return fmt.Errorf("update Spring Cloud App %q (Spring Cloud Service %q / Resource Group %q): %+v", name, serviceName, resourceGroup, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of Spring Cloud App %q (Spring Cloud Service %q / Resource Group %q): %+v", name, serviceName, resourceGroup, err)
	}

	d.SetId(resourceId)
	return resourceSpringCloudAppRead(d, meta)
}

func resourceSpringCloudAppUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.AppsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SpringCloudAppID(d.Id())
	if err != nil {
		return err
	}

	identity, err := expandSpringCloudAppIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return err
	}

	app := appplatform.AppResource{
		Identity: identity,
		Properties: &appplatform.AppResourceProperties{
			EnableEndToEndTLS: utils.Bool(d.Get("tls_enabled").(bool)),
			Public:            utils.Bool(d.Get("is_public").(bool)),
			HTTPSOnly:         utils.Bool(d.Get("https_only").(bool)),
			PersistentDisk:    expandSpringCloudAppPersistentDisk(d.Get("persistent_disk").([]interface{})),
		},
	}
	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.SpringName, id.AppName, app)
	if err != nil {
		return fmt.Errorf("update %s: %+v", id, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of %s: %+v", id, err)
	}

	return resourceSpringCloudAppRead(d, meta)
}

func resourceSpringCloudAppRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.AppsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SpringCloudAppID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.SpringName, id.AppName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Spring Cloud App %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading Spring Cloud App %q (Spring Cloud Service %q / Resource Group %q): %+v", id.AppName, id.SpringName, id.ResourceGroup, err)
	}

	d.Set("name", id.AppName)
	d.Set("service_name", id.SpringName)
	d.Set("resource_group_name", id.ResourceGroup)

	if err := d.Set("identity", flattenSpringCloudAppIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("setting `identity`: %s", err)
	}

	if prop := resp.Properties; prop != nil {
		d.Set("is_public", prop.Public)
		d.Set("https_only", prop.HTTPSOnly)
		d.Set("fqdn", prop.Fqdn)
		d.Set("url", prop.URL)
		d.Set("tls_enabled", prop.EnableEndToEndTLS)

		if err := d.Set("persistent_disk", flattenSpringCloudAppPersistentDisk(prop.PersistentDisk)); err != nil {
			return fmt.Errorf("setting `persistent_disk`: %s", err)
		}
	}

	return nil
}

func resourceSpringCloudAppDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.AppsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SpringCloudAppID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.SpringName, id.AppName); err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	return nil
}

func expandSpringCloudAppIdentity(input []interface{}) (*appplatform.ManagedIdentityProperties, error) {
	config, err := springCloudAppIdentity{}.Expand(input)
	if err != nil {
		return nil, err
	}

	return &appplatform.ManagedIdentityProperties{
		Type:        appplatform.ManagedIdentityType(config.Type),
		TenantID:    config.TenantId,
		PrincipalID: config.PrincipalId,
	}, nil
}

func expandSpringCloudAppPersistentDisk(input []interface{}) *appplatform.PersistentDisk {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	raw := input[0].(map[string]interface{})
	return &appplatform.PersistentDisk{
		SizeInGB:  utils.Int32(int32(raw["size_in_gb"].(int))),
		MountPath: utils.String(raw["mount_path"].(string)),
	}
}

func flattenSpringCloudAppIdentity(input *appplatform.ManagedIdentityProperties) []interface{} {
	var config *identity.ExpandedConfig
	if input != nil {
		config = &identity.ExpandedConfig{
			Type:        string(input.Type),
			PrincipalId: input.PrincipalID,
			TenantId:    input.TenantID,
		}
	}
	return springCloudAppIdentity{}.Flatten(config)
}

func flattenSpringCloudAppPersistentDisk(input *appplatform.PersistentDisk) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	sizeInGB := 0
	if input.SizeInGB != nil {
		sizeInGB = int(*input.SizeInGB)
	}

	mountPath := ""
	if input.MountPath != nil {
		mountPath = *input.MountPath
	}

	return []interface{}{
		map[string]interface{}{
			"size_in_gb": sizeInGB,
			"mount_path": mountPath,
		},
	}
}
