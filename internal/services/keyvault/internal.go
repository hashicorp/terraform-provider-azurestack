package keyvault

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/keyvault/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type deleteAndPurgeNestedItem interface {
	DeleteNestedItem(ctx context.Context) (autorest.Response, error)
	NestedItemHasBeenDeleted(ctx context.Context) (autorest.Response, error)

	PurgeNestedItem(ctx context.Context) (autorest.Response, error)
	NestedItemHasBeenPurged(ctx context.Context) (autorest.Response, error)
}

func deleteAndOptionallyPurge(ctx context.Context, description string, shouldPurge bool, helper deleteAndPurgeNestedItem) error {
	timeout, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context is missing a timeout")
	}

	log.Printf("[DEBUG] Deleting %s..", description)
	if resp, err := helper.DeleteNestedItem(ctx); err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("deleting %s: %+v", description, err)
	}
	log.Printf("[DEBUG] Waiting for %s to finish deleting..", description)
	stateConf := &resource.StateChangeConf{
		Pending: []string{"InProgress"},
		Target:  []string{"NotFound"},
		Refresh: func() (interface{}, string, error) {
			item, err := helper.NestedItemHasBeenDeleted(ctx)
			if err != nil {
				if utils.ResponseWasNotFound(item) {
					return item, "NotFound", nil
				}

				return nil, "Error", err
			}

			return item, "InProgress", nil
		},
		ContinuousTargetOccurence: 3,
		PollInterval:              5 * time.Second,
		Timeout:                   time.Until(timeout),
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for %s to be deleted: %+v", description, err)
	}
	log.Printf("[DEBUG] Deleted %s.", description)

	if !shouldPurge {
		log.Printf("[DEBUG] Skipping purging of %s as opted-out..", description)
		return nil
	}

	log.Printf("[DEBUG] Purging %s..", description)
	if _, err := helper.PurgeNestedItem(ctx); err != nil {
		return fmt.Errorf("purging %s: %+v", description, err)
	}

	log.Printf("[DEBUG] Waiting for %s to finish purging..", description)
	stateConf = &resource.StateChangeConf{
		Pending: []string{"InProgress"},
		Target:  []string{"NotFound"},
		Refresh: func() (interface{}, string, error) {
			item, err := helper.NestedItemHasBeenPurged(ctx)
			if err != nil {
				if utils.ResponseWasNotFound(item) {
					return item, "NotFound", nil
				}

				return nil, "Error", err
			}

			return item, "InProgress", nil
		},
		ContinuousTargetOccurence: 3,
		PollInterval:              5 * time.Second,
		Timeout:                   time.Until(timeout),
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for %s to finish purging: %+v", description, err)
	}
	log.Printf("[DEBUG] Purged %s.", description)

	return nil
}

func nestedItemResourceImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keyVaultsClient := meta.(*clients.Client).KeyVault
	resourcesClient := meta.(*clients.Client).Resource
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ParseNestedItemID(d.Id())
	if err != nil {
		return []*schema.ResourceData{d}, fmt.Errorf("parsing ID %q for Key Vault Child import: %v", d.Id(), err)
	}

	keyVaultId, err := keyVaultsClient.KeyVaultIDFromBaseUrl(ctx, resourcesClient, id.KeyVaultBaseUrl)
	if err != nil {
		return []*schema.ResourceData{d}, fmt.Errorf("retrieving the Resource ID the Key Vault at URL %q: %s", id.KeyVaultBaseUrl, err)
	}
	d.Set("key_vault_id", keyVaultId)

	return []*schema.ResourceData{d}, nil
}
