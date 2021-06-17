package containers

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2021-03-01/containerservice"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	computeValidate "github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/compute/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/containers/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
)

func SchemaDefaultNodePool() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				// Required
				"name": {
					Type:         pluginsdk.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validate.KubernetesAgentPoolName,
				},

				"type": {
					Type:     pluginsdk.TypeString,
					Optional: true,
					ForceNew: true,
					Default:  string(containerservice.AgentPoolTypeVirtualMachineScaleSets),
					ValidateFunc: validation.StringInSlice([]string{
						string(containerservice.AgentPoolTypeAvailabilitySet),
						string(containerservice.AgentPoolTypeVirtualMachineScaleSets),
					}, false),
				},

				"vm_size": {
					Type:         pluginsdk.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},

				// Optional
				"availability_zones": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &pluginsdk.Schema{
						Type: pluginsdk.TypeString,
					},
				},

				"enable_auto_scaling": {
					Type:     pluginsdk.TypeBool,
					Optional: true,
				},

				"enable_node_public_ip": {
					Type:     pluginsdk.TypeBool,
					Optional: true,
					ForceNew: true,
				},

				"enable_host_encryption": {
					Type:     pluginsdk.TypeBool,
					Optional: true,
					ForceNew: true,
				},

				"max_count": {
					Type:     pluginsdk.TypeInt,
					Optional: true,
					// NOTE: rather than setting `0` users should instead pass `null` here
					ValidateFunc: validation.IntBetween(1, 1000),
				},

				"max_pods": {
					Type:     pluginsdk.TypeInt,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},

				"min_count": {
					Type:     pluginsdk.TypeInt,
					Optional: true,
					// NOTE: rather than setting `0` users should instead pass `null` here
					ValidateFunc: validation.IntBetween(1, 1000),
				},

				"node_count": {
					Type:         pluginsdk.TypeInt,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.IntBetween(1, 1000),
				},

				"node_labels": {
					Type:     pluginsdk.TypeMap,
					ForceNew: true,
					Optional: true,
					Elem: &pluginsdk.Schema{
						Type: pluginsdk.TypeString,
					},
				},

				"node_public_ip_prefix_id": {
					Type:         pluginsdk.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: azure.ValidateResourceID,
					RequiredWith: []string{"default_node_pool.0.enable_node_public_ip"},
				},

				"node_taints": {
					Type:     pluginsdk.TypeList,
					ForceNew: true,
					Optional: true,
					Elem: &pluginsdk.Schema{
						Type: pluginsdk.TypeString,
					},
				},

				"tags": tags.Schema(),

				"os_disk_size_gb": {
					Type:         pluginsdk.TypeInt,
					Optional:     true,
					ForceNew:     true,
					Computed:     true,
					ValidateFunc: validation.IntAtLeast(1),
				},

				"os_disk_type": {
					Type:     pluginsdk.TypeString,
					Optional: true,
					ForceNew: true,
					Default:  containerservice.OSDiskTypeManaged,
					ValidateFunc: validation.StringInSlice([]string{
						string(containerservice.OSDiskTypeEphemeral),
						string(containerservice.OSDiskTypeManaged),
					}, false),
				},

				"vnet_subnet_id": {
					Type:         pluginsdk.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: azure.ValidateResourceID,
				},
				"orchestrator_version": {
					Type:         pluginsdk.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				"proximity_placement_group_id": {
					Type:         pluginsdk.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: computeValidate.ProximityPlacementGroupID,
				},
				"only_critical_addons_enabled": {
					Type:     pluginsdk.TypeBool,
					Optional: true,
					ForceNew: true,
				},

				"upgrade_settings": upgradeSettingsSchema(),
			},
		},
	}
}

func ConvertDefaultNodePoolToAgentPool(input *[]containerservice.ManagedClusterAgentPoolProfile) containerservice.AgentPool {
	defaultCluster := (*input)[0]
	return containerservice.AgentPool{
		Name: defaultCluster.Name,
		ManagedClusterAgentPoolProfileProperties: &containerservice.ManagedClusterAgentPoolProfileProperties{
			Count:                     defaultCluster.Count,
			VMSize:                    defaultCluster.VMSize,
			OsDiskSizeGB:              defaultCluster.OsDiskSizeGB,
			OsDiskType:                defaultCluster.OsDiskType,
			VnetSubnetID:              defaultCluster.VnetSubnetID,
			MaxPods:                   defaultCluster.MaxPods,
			OsType:                    defaultCluster.OsType,
			MaxCount:                  defaultCluster.MaxCount,
			MinCount:                  defaultCluster.MinCount,
			EnableAutoScaling:         defaultCluster.EnableAutoScaling,
			Type:                      defaultCluster.Type,
			OrchestratorVersion:       defaultCluster.OrchestratorVersion,
			ProximityPlacementGroupID: defaultCluster.ProximityPlacementGroupID,
			AvailabilityZones:         defaultCluster.AvailabilityZones,
			EnableNodePublicIP:        defaultCluster.EnableNodePublicIP,
			NodePublicIPPrefixID:      defaultCluster.NodePublicIPPrefixID,
			ScaleSetPriority:          defaultCluster.ScaleSetPriority,
			ScaleSetEvictionPolicy:    defaultCluster.ScaleSetEvictionPolicy,
			SpotMaxPrice:              defaultCluster.SpotMaxPrice,
			Mode:                      defaultCluster.Mode,
			NodeLabels:                defaultCluster.NodeLabels,
			NodeTaints:                defaultCluster.NodeTaints,
			Tags:                      defaultCluster.Tags,
			UpgradeSettings:           defaultCluster.UpgradeSettings,
		},
	}
}

func ExpandDefaultNodePool(d *pluginsdk.ResourceData) (*[]containerservice.ManagedClusterAgentPoolProfile, error) {
	input := d.Get("default_node_pool").([]interface{})

	raw := input[0].(map[string]interface{})
	enableAutoScaling := raw["enable_auto_scaling"].(bool)
	nodeLabelsRaw := raw["node_labels"].(map[string]interface{})
	nodeLabels := utils.ExpandMapStringPtrString(nodeLabelsRaw)
	nodeTaintsRaw := raw["node_taints"].([]interface{})
	nodeTaints := utils.ExpandStringSlice(nodeTaintsRaw)

	if len(*nodeTaints) != 0 {
		return nil, fmt.Errorf("The AKS API has removed support for tainting all nodes in the default node pool and it is no longer possible to configure this. To taint a node pool, create a separate one.")
	}

	criticalAddonsEnabled := raw["only_critical_addons_enabled"].(bool)
	if criticalAddonsEnabled {
		*nodeTaints = append(*nodeTaints, "CriticalAddonsOnly=true:NoSchedule")
	}

	t := raw["tags"].(map[string]interface{})

	profile := containerservice.ManagedClusterAgentPoolProfile{
		EnableAutoScaling:      utils.Bool(enableAutoScaling),
		EnableNodePublicIP:     utils.Bool(raw["enable_node_public_ip"].(bool)),
		EnableEncryptionAtHost: utils.Bool(raw["enable_host_encryption"].(bool)),
		Name:                   utils.String(raw["name"].(string)),
		NodeLabels:             nodeLabels,
		NodeTaints:             nodeTaints,
		Tags:                   tags.Expand(t),
		Type:                   containerservice.AgentPoolType(raw["type"].(string)),
		VMSize:                 utils.String(raw["vm_size"].(string)),

		// at this time the default node pool has to be Linux or the AKS cluster fails to provision with:
		// Pods not in Running status: coredns-7fc597cc45-v5z7x,coredns-autoscaler-7ccc76bfbd-djl7j,metrics-server-cbd95f966-5rl97,tunnelfront-7d9884977b-wpbvn
		// Windows agents can be configured via the separate node pool resource
		OsType: containerservice.OSTypeLinux,

		// without this set the API returns:
		// Code="MustDefineAtLeastOneSystemPool" Message="Must define at least one system pool."
		// since this is the "default" node pool we can assume this is a system node pool
		Mode: containerservice.AgentPoolModeSystem,

		UpgradeSettings: expandUpgradeSettings(raw["upgrade_settings"].([]interface{})),

		// // TODO: support these in time
		// ScaleSetEvictionPolicy: "",
		// ScaleSetPriority:       "",
	}

	availabilityZonesRaw := raw["availability_zones"].([]interface{})
	availabilityZones := utils.ExpandStringSlice(availabilityZonesRaw)

	// otherwise: Standard Load Balancer is required for availability zone.
	if len(*availabilityZones) > 0 {
		profile.AvailabilityZones = availabilityZones
	}

	if maxPods := int32(raw["max_pods"].(int)); maxPods > 0 {
		profile.MaxPods = utils.Int32(maxPods)
	}

	if prefixID := raw["node_public_ip_prefix_id"].(string); prefixID != "" {
		profile.NodePublicIPPrefixID = utils.String(prefixID)
	}

	if osDiskSizeGB := int32(raw["os_disk_size_gb"].(int)); osDiskSizeGB > 0 {
		profile.OsDiskSizeGB = utils.Int32(osDiskSizeGB)
	}

	profile.OsDiskType = containerservice.OSDiskTypeManaged
	if osDiskType := raw["os_disk_type"].(string); osDiskType != "" {
		profile.OsDiskType = containerservice.OSDiskType(raw["os_disk_type"].(string))
	}

	if vnetSubnetID := raw["vnet_subnet_id"].(string); vnetSubnetID != "" {
		profile.VnetSubnetID = utils.String(vnetSubnetID)
	}

	if orchestratorVersion := raw["orchestrator_version"].(string); orchestratorVersion != "" {
		profile.OrchestratorVersion = utils.String(orchestratorVersion)
	}

	if proximityPlacementGroupId := raw["proximity_placement_group_id"].(string); proximityPlacementGroupId != "" {
		profile.ProximityPlacementGroupID = utils.String(proximityPlacementGroupId)
	}

	count := raw["node_count"].(int)
	maxCount := raw["max_count"].(int)
	minCount := raw["min_count"].(int)

	// Count must always be set (see #6094), RP behaviour has changed
	// since the API version upgrade in v2.1.0 making Count required
	// for all create/update requests
	profile.Count = utils.Int32(int32(count))

	if enableAutoScaling {
		// if Count has not been set use min count
		if count == 0 {
			count = minCount
			profile.Count = utils.Int32(int32(count))
		}

		// Count must be set for the initial creation when using AutoScaling but cannot be updated
		if d.HasChange("default_node_pool.0.node_count") && !d.IsNewResource() {
			return nil, fmt.Errorf("cannot change `node_count` when `enable_auto_scaling` is set to `true`")
		}

		if maxCount > 0 {
			profile.MaxCount = utils.Int32(int32(maxCount))
			if maxCount < count {
				return nil, fmt.Errorf("`node_count`(%d) must be equal to or less than `max_count`(%d) when `enable_auto_scaling` is set to `true`", count, maxCount)
			}
		} else {
			return nil, fmt.Errorf("`max_count` must be configured when `enable_auto_scaling` is set to `true`")
		}

		if minCount > 0 {
			profile.MinCount = utils.Int32(int32(minCount))

			if minCount > count && d.IsNewResource() {
				return nil, fmt.Errorf("`node_count`(%d) must be equal to or greater than `min_count`(%d) when `enable_auto_scaling` is set to `true`", count, minCount)
			}
		} else {
			return nil, fmt.Errorf("`min_count` must be configured when `enable_auto_scaling` is set to `true`")
		}

		if minCount > maxCount {
			return nil, fmt.Errorf("`max_count` must be >= `min_count`")
		}
	} else if minCount > 0 || maxCount > 0 {
		return nil, fmt.Errorf("`max_count`(%d) and `min_count`(%d) must be set to `null` when `enable_auto_scaling` is set to `false`", maxCount, minCount)
	}

	return &[]containerservice.ManagedClusterAgentPoolProfile{
		profile,
	}, nil
}

func FlattenDefaultNodePool(input *[]containerservice.ManagedClusterAgentPoolProfile, d *pluginsdk.ResourceData) (*[]interface{}, error) {
	if input == nil {
		return &[]interface{}{}, nil
	}

	agentPool, err := findDefaultNodePool(input, d)
	if err != nil {
		return nil, err
	}

	var availabilityZones []string
	if agentPool.AvailabilityZones != nil {
		availabilityZones = *agentPool.AvailabilityZones
	}

	count := 0
	if agentPool.Count != nil {
		count = int(*agentPool.Count)
	}

	enableAutoScaling := false
	if agentPool.EnableAutoScaling != nil {
		enableAutoScaling = *agentPool.EnableAutoScaling
	}

	enableNodePublicIP := false
	if agentPool.EnableNodePublicIP != nil {
		enableNodePublicIP = *agentPool.EnableNodePublicIP
	}

	enableHostEncryption := false
	if agentPool.EnableEncryptionAtHost != nil {
		enableHostEncryption = *agentPool.EnableEncryptionAtHost
	}

	maxCount := 0
	if agentPool.MaxCount != nil {
		maxCount = int(*agentPool.MaxCount)
	}

	maxPods := 0
	if agentPool.MaxPods != nil {
		maxPods = int(*agentPool.MaxPods)
	}

	minCount := 0
	if agentPool.MinCount != nil {
		minCount = int(*agentPool.MinCount)
	}

	name := ""
	if agentPool.Name != nil {
		name = *agentPool.Name
	}

	var nodeLabels map[string]string
	if agentPool.NodeLabels != nil {
		nodeLabels = make(map[string]string)
		for k, v := range agentPool.NodeLabels {
			nodeLabels[k] = *v
		}
	}

	nodePublicIPPrefixID := ""
	if agentPool.NodePublicIPPrefixID != nil {
		nodePublicIPPrefixID = *agentPool.NodePublicIPPrefixID
	}

	criticalAddonsEnabled := false
	if agentPool.NodeTaints != nil {
		for _, taint := range *agentPool.NodeTaints {
			if strings.EqualFold(taint, "CriticalAddonsOnly=true:NoSchedule") {
				criticalAddonsEnabled = true
			}
		}
	}

	osDiskSizeGB := 0
	if agentPool.OsDiskSizeGB != nil {
		osDiskSizeGB = int(*agentPool.OsDiskSizeGB)
	}

	osDiskType := containerservice.OSDiskTypeManaged
	if agentPool.OsDiskType != "" {
		osDiskType = agentPool.OsDiskType
	}

	vnetSubnetId := ""
	if agentPool.VnetSubnetID != nil {
		vnetSubnetId = *agentPool.VnetSubnetID
	}

	orchestratorVersion := ""
	if agentPool.OrchestratorVersion != nil {
		orchestratorVersion = *agentPool.OrchestratorVersion
	}

	proximityPlacementGroupId := ""
	if agentPool.ProximityPlacementGroupID != nil {
		proximityPlacementGroupId = *agentPool.ProximityPlacementGroupID
	}

	vmSize := ""
	if agentPool.VMSize != nil {
		vmSize = *agentPool.VMSize
	}

	upgradeSettings := flattenUpgradeSettings(agentPool.UpgradeSettings)

	return &[]interface{}{
		map[string]interface{}{
			"availability_zones":           availabilityZones,
			"enable_auto_scaling":          enableAutoScaling,
			"enable_node_public_ip":        enableNodePublicIP,
			"enable_host_encryption":       enableHostEncryption,
			"max_count":                    maxCount,
			"max_pods":                     maxPods,
			"min_count":                    minCount,
			"name":                         name,
			"node_count":                   count,
			"node_labels":                  nodeLabels,
			"node_public_ip_prefix_id":     nodePublicIPPrefixID,
			"node_taints":                  []string{},
			"os_disk_size_gb":              osDiskSizeGB,
			"os_disk_type":                 string(osDiskType),
			"tags":                         tags.Flatten(agentPool.Tags),
			"type":                         string(agentPool.Type),
			"vm_size":                      vmSize,
			"orchestrator_version":         orchestratorVersion,
			"proximity_placement_group_id": proximityPlacementGroupId,
			"upgrade_settings":             upgradeSettings,
			"vnet_subnet_id":               vnetSubnetId,
			"only_critical_addons_enabled": criticalAddonsEnabled,
		},
	}, nil
}

func findDefaultNodePool(input *[]containerservice.ManagedClusterAgentPoolProfile, d *pluginsdk.ResourceData) (*containerservice.ManagedClusterAgentPoolProfile, error) {
	// first try loading this from the Resource Data if possible (e.g. when Created)
	defaultNodePoolName := d.Get("default_node_pool.0.name")

	var agentPool *containerservice.ManagedClusterAgentPoolProfile
	if defaultNodePoolName != "" {
		// find it
		for _, v := range *input {
			if v.Name != nil && *v.Name == defaultNodePoolName {
				agentPool = &v
				break
			}
		}
	}

	if agentPool == nil {
		// otherwise we need to fall back to the name of the first agent pool
		for _, v := range *input {
			if v.Name == nil {
				continue
			}
			if v.Mode != containerservice.AgentPoolModeSystem {
				continue
			}

			defaultNodePoolName = *v.Name
			agentPool = &v
			break
		}

		if defaultNodePoolName == nil {
			return nil, fmt.Errorf("Unable to Determine Default Agent Pool")
		}
	}

	if agentPool == nil {
		return nil, fmt.Errorf("The Default Agent Pool %q was not found", defaultNodePoolName)
	}

	return agentPool, nil
}
