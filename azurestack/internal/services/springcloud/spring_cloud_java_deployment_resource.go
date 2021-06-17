package springcloud

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/appplatform/mgmt/2020-11-01-preview/appplatform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/springcloud/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/springcloud/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceSpringCloudJavaDeployment() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceSpringCloudJavaDeploymentCreate,
		Read:   resourceSpringCloudJavaDeploymentRead,
		Update: resourceSpringCloudJavaDeploymentUpdate,
		Delete: resourceSpringCloudJavaDeploymentDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SpringCloudDeploymentID(id)
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
				ValidateFunc: validate.SpringCloudDeploymentName,
			},

			"spring_cloud_app_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SpringCloudAppID,
			},

			"cpu": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 4),
			},

			"environment_variables": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"instance_count": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 500),
			},

			"jvm_options": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"memory_in_gb": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 8),
			},

			"runtime_version": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(appplatform.Java8),
					string(appplatform.Java11),
				}, false),
				Default: string(appplatform.Java8),
			},
		},
	}
}

func resourceSpringCloudJavaDeploymentCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.DeploymentsClient
	servicesClient := meta.(*clients.Client).AppPlatform.ServicesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	appId, err := parse.SpringCloudAppID(d.Get("spring_cloud_app_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewSpringCloudDeploymentID(subscriptionId, appId.ResourceGroup, appId.SpringName, appId.AppName, d.Get("name").(string))
	existing, err := client.Get(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.DeploymentName)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_spring_cloud_java_deployment", id.ID())
	}

	service, err := servicesClient.Get(ctx, appId.ResourceGroup, appId.SpringName)
	if err != nil {
		return fmt.Errorf("checking for presence of existing Spring Cloud Service %q (Resource Group %q): %+v", appId.SpringName, appId.ResourceGroup, err)
	}
	if service.Sku == nil || service.Sku.Name == nil || service.Sku.Tier == nil {
		return fmt.Errorf("invalid `sku` for Spring Cloud Service %q (Resource Group %q)", appId.SpringName, appId.ResourceGroup)
	}

	deployment := appplatform.DeploymentResource{
		Sku: &appplatform.Sku{
			Name:     service.Sku.Name,
			Tier:     service.Sku.Tier,
			Capacity: utils.Int32(int32(d.Get("instance_count").(int))),
		},
		Properties: &appplatform.DeploymentResourceProperties{
			Source: &appplatform.UserSourceInfo{
				Type:         appplatform.Jar,
				RelativePath: utils.String("<default>"),
			},
			DeploymentSettings: &appplatform.DeploymentSettings{
				CPU:                  utils.Int32(int32(d.Get("cpu").(int))),
				MemoryInGB:           utils.Int32(int32(d.Get("memory_in_gb").(int))),
				JvmOptions:           utils.String(d.Get("jvm_options").(string)),
				EnvironmentVariables: expandSpringCloudDeploymentEnvironmentVariables(d.Get("environment_variables").(map[string]interface{})),
				RuntimeVersion:       appplatform.RuntimeVersion(d.Get("runtime_version").(string)),
			},
		},
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.DeploymentName, deployment)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceSpringCloudJavaDeploymentRead(d, meta)
}

func resourceSpringCloudJavaDeploymentUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.DeploymentsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SpringCloudDeploymentID(d.Id())
	if err != nil {
		return err
	}

	existing, err := client.Get(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.DeploymentName)
	if err != nil {
		return fmt.Errorf("reading existing %s: %+v", id, err)
	}
	if existing.Sku == nil || existing.Properties == nil || existing.Properties.DeploymentSettings == nil {
		return fmt.Errorf("nil `sku`, `properties` or `properties.deploymentSettings` for %s: %+v", id, err)
	}

	if d.HasChange("instance_count") {
		existing.Sku.Capacity = utils.Int32(int32(d.Get("instance_count").(int)))
	}

	if d.HasChange("cpu") {
		existing.Properties.DeploymentSettings.CPU = utils.Int32(int32(d.Get("cpu").(int)))
	}

	if d.HasChange("environment_variables") {
		existing.Properties.DeploymentSettings.EnvironmentVariables = expandSpringCloudDeploymentEnvironmentVariables(d.Get("environment_variables").(map[string]interface{}))
	}

	if d.HasChange("jvm_options") {
		existing.Properties.DeploymentSettings.JvmOptions = utils.String(d.Get("jvm_options").(string))
	}

	if d.HasChange("memory_in_gb") {
		existing.Properties.DeploymentSettings.MemoryInGB = utils.Int32(int32(d.Get("memory_in_gb").(int)))
	}

	if d.HasChange("runtime_version") {
		existing.Properties.DeploymentSettings.RuntimeVersion = appplatform.RuntimeVersion(d.Get("runtime_version").(string))
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.DeploymentName, existing)
	if err != nil {
		return fmt.Errorf("updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of %s: %+v", id, err)
	}

	return resourceSpringCloudJavaDeploymentRead(d, meta)
}

func resourceSpringCloudJavaDeploymentRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.DeploymentsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SpringCloudDeploymentID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.DeploymentName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Spring Cloud deployment %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading Spring Cloud Deployment %q (Spring Cloud Service %q / App %q / resource Group %q): %+v", id.DeploymentName, id.SpringName, id.AppName, id.ResourceGroup, err)
	}

	d.Set("name", id.DeploymentName)
	d.Set("spring_cloud_app_id", parse.NewSpringCloudAppID(id.SubscriptionId, id.ResourceGroup, id.SpringName, id.AppName).ID())
	if resp.Sku != nil {
		d.Set("instance_count", resp.Sku.Capacity)
	}
	if resp.Properties != nil && resp.Properties.DeploymentSettings != nil {
		settings := resp.Properties.DeploymentSettings
		d.Set("cpu", settings.CPU)
		d.Set("memory_in_gb", settings.MemoryInGB)
		d.Set("jvm_options", settings.JvmOptions)
		d.Set("environment_variables", flattenSpringCloudDeploymentEnvironmentVariables(settings.EnvironmentVariables))
		d.Set("runtime_version", settings.RuntimeVersion)
	}

	return nil
}

func resourceSpringCloudJavaDeploymentDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.DeploymentsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SpringCloudDeploymentID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.DeploymentName); err != nil {
		return fmt.Errorf("deleting Spring Cloud Deployment %q (Spring Cloud Service %q / App %q / resource Group %q): %+v", id.DeploymentName, id.SpringName, id.AppName, id.ResourceGroup, err)
	}

	return nil
}

func expandSpringCloudDeploymentEnvironmentVariables(envMap map[string]interface{}) map[string]*string {
	output := make(map[string]*string, len(envMap))

	for k, v := range envMap {
		output[k] = utils.String(v.(string))
	}

	return output
}

func flattenSpringCloudDeploymentEnvironmentVariables(envMap map[string]*string) map[string]interface{} {
	output := make(map[string]interface{}, len(envMap))
	for i, v := range envMap {
		if v == nil {
			continue
		}
		output[i] = *v
	}
	return output
}
