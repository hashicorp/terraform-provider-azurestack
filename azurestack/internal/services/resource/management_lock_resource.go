package resource

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-09-01/locks"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/resource/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceManagementLock() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceManagementLockCreateUpdate,
		Read:   resourceManagementLockRead,
		Delete: resourceManagementLockDelete,
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
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ManagementLockName,
			},

			"scope": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"lock_level": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(locks.CanNotDelete),
					string(locks.ReadOnly),
				}, false),
			},

			"notes": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
		},
	}
}

func resourceManagementLockCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.LocksClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for AzureRM Management Lock creation.")

	name := d.Get("name").(string)
	scope := d.Get("scope").(string)

	if d.IsNewResource() {
		existing, err := client.GetByScope(ctx, scope, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Management Lock %q (Scope %q): %s", name, scope, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_management_lock", *existing.ID)
		}
	}

	lockLevel := d.Get("lock_level").(string)
	notes := d.Get("notes").(string)

	lock := locks.ManagementLockObject{
		ManagementLockProperties: &locks.ManagementLockProperties{
			Level: locks.LockLevel(lockLevel),
			Notes: utils.String(notes),
		},
	}

	if _, err := client.CreateOrUpdateByScope(ctx, scope, name, lock); err != nil {
		return err
	}

	read, err := client.GetByScope(ctx, scope, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read ID of AzureRM Management Lock %q (Scope %q)", name, scope)
	}

	d.SetId(*read.ID)
	return resourceManagementLockRead(d, meta)
}

func resourceManagementLockRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.LocksClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := ParseAzureRMLockId(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetByScope(ctx, id.Scope, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on AzureRM Management Lock %q (Scope %q): %+v", id.Name, id.Scope, err)
	}

	d.Set("name", resp.Name)
	d.Set("scope", id.Scope)

	if props := resp.ManagementLockProperties; props != nil {
		d.Set("lock_level", string(props.Level))
		d.Set("notes", props.Notes)
	}

	return nil
}

func resourceManagementLockDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.LocksClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := ParseAzureRMLockId(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.DeleteByScope(ctx, id.Scope, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("Error issuing AzureRM delete request for Management Lock %q (Scope %q): %+v", id.Name, id.Scope, err)
	}

	return nil
}

type AzureManagementLockId struct {
	Scope string
	Name  string
}

func ParseAzureRMLockId(id string) (*AzureManagementLockId, error) {
	segments := strings.Split(id, "/providers/Microsoft.Authorization/locks/")
	if len(segments) != 2 {
		return nil, fmt.Errorf("Expected ID to be in the format `{scope}/providers/Microsoft.Authorization/locks/{name} - got %d segments", len(segments))
	}

	scope := segments[0]
	name := segments[1]
	lockId := AzureManagementLockId{
		Scope: scope,
		Name:  name,
	}
	return &lockId, nil
}
