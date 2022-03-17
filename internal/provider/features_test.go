package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/features"
)

func TestExpandFeatures(t *testing.T) {
	testData := []struct {
		Name     string
		Input    []interface{}
		EnvVars  map[string]interface{}
		Expected features.UserFeatures
	}{
		{
			Name:  "Empty Block",
			Input: []interface{}{},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
			},
		},
		{
			Name: "Complete Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"key_vault": []interface{}{
						map[string]interface{}{
							"purge_soft_delete_on_destroy":    true,
							"recover_soft_deleted_key_vaults": true,
						},
					},
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": true,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: true,
				},
			},
		},
		{
			Name: "Complete Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"key_vault": []interface{}{
						map[string]interface{}{
							"purge_soft_delete_on_destroy":    false,
							"recover_soft_deleted_key_vaults": false,
						},
					},
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Logf("[DEBUG] Test Case: %q", testCase.Name)
		result := expandFeatures(testCase.Input)
		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Fatalf("Expected %+v but got %+v", result, testCase.Expected)
		}
	}
}

func TestExpandFeaturesResourceGroup(t *testing.T) {
	testData := []struct {
		Name     string
		Input    []interface{}
		EnvVars  map[string]interface{}
		Expected features.UserFeatures
	}{
		{
			Name: "Empty Block",
			Input: []interface{}{
				map[string]interface{}{
					"resource_group": []interface{}{},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
			},
		},
		{
			Name: "Prevent Deletion If Contains Resources Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": true,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: true,
				},
			},
		},
		{
			Name: "Prevent Deletion If Contains Resources Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Logf("[DEBUG] Test Case: %q", testCase.Name)
		result := expandFeatures(testCase.Input)
		if !reflect.DeepEqual(result.ResourceGroup, testCase.Expected.ResourceGroup) {
			t.Fatalf("Expected %+v but got %+v", result.ResourceGroup, testCase.Expected.ResourceGroup)
		}
	}
}
