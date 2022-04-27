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
				VirtualMachine: features.VirtualMachineFeatures{
					DeleteOSDiskOnDeletion:     true,
					GracefulShutdown:           false,
					SkipShutdownAndForceDelete: false,
				},
				VirtualMachineScaleSet: features.VirtualMachineScaleSetFeatures{
					ForceDelete:               false,
					RollInstancesWhenRequired: true,
					ScaleToZeroOnDelete:       true,
				},
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
			},
		},
		{
			Name: "Complete Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": true,
						},
					},
					"template_deployment": []interface{}{
						map[string]interface{}{
							"delete_nested_items_during_deletion": true,
						},
					},
					"virtual_machine": []interface{}{
						map[string]interface{}{
							"delete_os_disk_on_deletion":     true,
							"graceful_shutdown":              true,
							"skip_shutdown_and_force_delete": true,
						},
					},
					"virtual_machine_scale_set": []interface{}{
						map[string]interface{}{
							"roll_instances_when_required":  true,
							"force_delete":                  true,
							"scale_to_zero_before_deletion": true,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: true,
				},
				VirtualMachine: features.VirtualMachineFeatures{
					DeleteOSDiskOnDeletion:     true,
					GracefulShutdown:           true,
					SkipShutdownAndForceDelete: true,
				},
				VirtualMachineScaleSet: features.VirtualMachineScaleSetFeatures{
					RollInstancesWhenRequired: true,
					ForceDelete:               true,
					ScaleToZeroOnDelete:       true,
				},
			},
		},
		{
			Name: "Complete Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"resource_group": []interface{}{
						map[string]interface{}{
							"prevent_deletion_if_contains_resources": false,
						},
					},
					"template_deployment": []interface{}{
						map[string]interface{}{
							"delete_nested_items_during_deletion": false,
						},
					},
					"virtual_machine": []interface{}{
						map[string]interface{}{
							"delete_os_disk_on_deletion":     false,
							"graceful_shutdown":              false,
							"skip_shutdown_and_force_delete": false,
						},
					},
					"virtual_machine_scale_set": []interface{}{
						map[string]interface{}{
							"force_delete":                  false,
							"roll_instances_when_required":  false,
							"scale_to_zero_before_deletion": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				ResourceGroup: features.ResourceGroupFeatures{
					PreventDeletionIfContainsResources: false,
				},
				VirtualMachine: features.VirtualMachineFeatures{
					DeleteOSDiskOnDeletion:     false,
					GracefulShutdown:           false,
					SkipShutdownAndForceDelete: false,
				},
				VirtualMachineScaleSet: features.VirtualMachineScaleSetFeatures{
					ForceDelete:               false,
					RollInstancesWhenRequired: false,
					ScaleToZeroOnDelete:       false,
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

func TestExpandFeaturesVirtualMachine(t *testing.T) {
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
					"virtual_machine": []interface{}{},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachine: features.VirtualMachineFeatures{
					DeleteOSDiskOnDeletion:     true,
					GracefulShutdown:           false,
					SkipShutdownAndForceDelete: false,
				},
			},
		},
		{
			Name: "Delete OS Disk Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"virtual_machine": []interface{}{
						map[string]interface{}{
							"delete_os_disk_on_deletion": true,
							"graceful_shutdown":          false,
							"force_delete":               false,
							"shutdown_before_deletion":   false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachine: features.VirtualMachineFeatures{
					DeleteOSDiskOnDeletion:     true,
					GracefulShutdown:           false,
					SkipShutdownAndForceDelete: false,
				},
			},
		},
		{
			Name: "Graceful Shutdown Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"virtual_machine": []interface{}{
						map[string]interface{}{
							"delete_os_disk_on_deletion": false,
							"graceful_shutdown":          true,
							"force_delete":               false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachine: features.VirtualMachineFeatures{
					DeleteOSDiskOnDeletion:     false,
					GracefulShutdown:           true,
					SkipShutdownAndForceDelete: false,
				},
			},
		},
		{
			Name: "Skip Shutdown and Force Delete Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"virtual_machine": []interface{}{
						map[string]interface{}{
							"delete_os_disk_on_deletion":     false,
							"graceful_shutdown":              false,
							"skip_shutdown_and_force_delete": true,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachine: features.VirtualMachineFeatures{
					DeleteOSDiskOnDeletion:     false,
					GracefulShutdown:           false,
					SkipShutdownAndForceDelete: true,
				},
			},
		},
		{
			Name: "All Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"virtual_machine": []interface{}{
						map[string]interface{}{
							"delete_os_disk_on_deletion":     false,
							"graceful_shutdown":              false,
							"skip_shutdown_and_force_delete": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachine: features.VirtualMachineFeatures{
					DeleteOSDiskOnDeletion:     false,
					GracefulShutdown:           false,
					SkipShutdownAndForceDelete: false,
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Logf("[DEBUG] Test Case: %q", testCase.Name)
		result := expandFeatures(testCase.Input)
		if !reflect.DeepEqual(result.VirtualMachine, testCase.Expected.VirtualMachine) {
			t.Fatalf("Expected %+v but got %+v", result.VirtualMachine, testCase.Expected.VirtualMachine)
		}
	}
}

func TestExpandFeaturesVirtualMachineScaleSet(t *testing.T) {
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
					"virtual_machine_scale_set": []interface{}{},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachineScaleSet: features.VirtualMachineScaleSetFeatures{
					RollInstancesWhenRequired: true,
					ScaleToZeroOnDelete:       true,
				},
			},
		},
		{
			Name: "Force Delete Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"virtual_machine_scale_set": []interface{}{
						map[string]interface{}{
							"force_delete":                 true,
							"roll_instances_when_required": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachineScaleSet: features.VirtualMachineScaleSetFeatures{
					ForceDelete:               true,
					RollInstancesWhenRequired: false,
					ScaleToZeroOnDelete:       true,
				},
			},
		},
		{
			Name: "Roll Instances Enabled",
			Input: []interface{}{
				map[string]interface{}{
					"virtual_machine_scale_set": []interface{}{
						map[string]interface{}{
							"force_delete":                 false,
							"roll_instances_when_required": true,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachineScaleSet: features.VirtualMachineScaleSetFeatures{
					ForceDelete:               false,
					RollInstancesWhenRequired: true,
					ScaleToZeroOnDelete:       true,
				},
			},
		},
		{
			Name: "Scale In On Delete Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"virtual_machine_scale_set": []interface{}{
						map[string]interface{}{
							"force_delete":                  false,
							"roll_instances_when_required":  true,
							"scale_to_zero_before_deletion": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachineScaleSet: features.VirtualMachineScaleSetFeatures{
					ForceDelete:               false,
					RollInstancesWhenRequired: true,
					ScaleToZeroOnDelete:       false,
				},
			},
		},
		{
			Name: "All Fields Disabled",
			Input: []interface{}{
				map[string]interface{}{
					"virtual_machine_scale_set": []interface{}{
						map[string]interface{}{
							"force_delete":                  false,
							"roll_instances_when_required":  false,
							"scale_to_zero_before_deletion": false,
						},
					},
				},
			},
			Expected: features.UserFeatures{
				VirtualMachineScaleSet: features.VirtualMachineScaleSetFeatures{
					ForceDelete:               false,
					RollInstancesWhenRequired: false,
					ScaleToZeroOnDelete:       false,
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Logf("[DEBUG] Test Case: %q", testCase.Name)
		result := expandFeatures(testCase.Input)
		if !reflect.DeepEqual(result.VirtualMachineScaleSet, testCase.Expected.VirtualMachineScaleSet) {
			t.Fatalf("Expected %+v but got %+v", testCase.Expected.VirtualMachineScaleSet, result.VirtualMachineScaleSet)
		}
	}
}
