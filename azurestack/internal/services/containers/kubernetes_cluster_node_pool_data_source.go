package containers

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2021-03-01/containerservice"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/containers/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func dataSourceKubernetesClusterNodePool() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceKubernetesClusterNodePoolRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.KubernetesAgentPoolName,
			},

			"kubernetes_cluster_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			// Computed
			"availability_zones": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"enable_auto_scaling": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"enable_node_public_ip": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"eviction_policy": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"max_count": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"max_pods": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"mode": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"min_count": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"node_count": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"node_labels": {
				Type:     pluginsdk.TypeMap,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"node_public_ip_prefix_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"node_taints": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"orchestrator_version": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"os_disk_size_gb": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"os_disk_type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"os_type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"priority": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"proximity_placement_group_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"spot_max_price": {
				Type:     pluginsdk.TypeFloat,
				Computed: true,
			},

			"tags": tags.SchemaDataSource(),

			"upgrade_settings": upgradeSettingsForDataSourceSchema(),

			"vm_size": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"vnet_subnet_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKubernetesClusterNodePoolRead(d *pluginsdk.ResourceData, meta interface{}) error {
	clustersClient := meta.(*clients.Client).Containers.KubernetesClustersClient
	poolsClient := meta.(*clients.Client).Containers.AgentPoolsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	nodePoolName := d.Get("name").(string)
	clusterName := d.Get("kubernetes_cluster_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	// if the parent cluster doesn't exist then the node pool won't
	cluster, err := clustersClient.Get(ctx, resourceGroup, clusterName)
	if err != nil {
		if utils.ResponseWasNotFound(cluster.Response) {
			return fmt.Errorf("Kubernetes Cluster %q was not found in Resource Group %q", clusterName, resourceGroup)
		}

		return fmt.Errorf("retrieving Managed Kubernetes Cluster %q (Resource Group %q): %+v", clusterName, resourceGroup, err)
	}

	resp, err := poolsClient.Get(ctx, resourceGroup, clusterName, nodePoolName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Node Pool %q was not found in Managed Kubernetes Cluster %q / Resource Group %q", nodePoolName, clusterName, resourceGroup)
		}

		return fmt.Errorf("retrieving Node Pool %q (Managed Kubernetes Cluster %q / Resource Group %q): %+v", nodePoolName, clusterName, resourceGroup, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("retrieving Node Pool %q (Managed Kubernetes Cluster %q / Resource Group %q): `id` was nil", nodePoolName, clusterName, resourceGroup)
	}

	d.SetId(*resp.ID)
	d.Set("name", nodePoolName)
	d.Set("kubernetes_cluster_name", clusterName)
	d.Set("resource_group_name", resourceGroup)

	if props := resp.ManagedClusterAgentPoolProfileProperties; props != nil {
		if err := d.Set("availability_zones", utils.FlattenStringSlice(props.AvailabilityZones)); err != nil {
			return fmt.Errorf("setting `availability_zones`: %+v", err)
		}

		d.Set("enable_auto_scaling", props.EnableAutoScaling)
		d.Set("enable_node_public_ip", props.EnableNodePublicIP)

		evictionPolicy := ""
		if props.ScaleSetEvictionPolicy != "" {
			evictionPolicy = string(props.ScaleSetEvictionPolicy)
		}
		d.Set("eviction_policy", evictionPolicy)

		maxCount := 0
		if props.MaxCount != nil {
			maxCount = int(*props.MaxCount)
		}
		d.Set("max_count", maxCount)

		maxPods := 0
		if props.MaxPods != nil {
			maxPods = int(*props.MaxPods)
		}
		d.Set("max_pods", maxPods)

		minCount := 0
		if props.MinCount != nil {
			minCount = int(*props.MinCount)
		}
		d.Set("min_count", minCount)

		mode := string(containerservice.AgentPoolModeUser)
		if props.Mode != "" {
			mode = string(props.Mode)
		}
		d.Set("mode", mode)

		count := 0
		if props.Count != nil {
			count = int(*props.Count)
		}
		d.Set("node_count", count)

		if err := d.Set("node_labels", props.NodeLabels); err != nil {
			return fmt.Errorf("setting `node_labels`: %+v", err)
		}

		d.Set("node_public_ip_prefix_id", props.NodePublicIPPrefixID)

		if err := d.Set("node_taints", utils.FlattenStringSlice(props.NodeTaints)); err != nil {
			return fmt.Errorf("setting `node_taints`: %+v", err)
		}

		d.Set("orchestrator_version", props.OrchestratorVersion)
		osDiskSizeGB := 0
		if props.OsDiskSizeGB != nil {
			osDiskSizeGB = int(*props.OsDiskSizeGB)
		}
		d.Set("os_disk_size_gb", osDiskSizeGB)

		osDiskType := containerservice.OSDiskTypeManaged
		if props.OsDiskType != "" {
			osDiskType = props.OsDiskType
		}
		d.Set("os_disk_type", string(osDiskType))
		d.Set("os_type", string(props.OsType))

		// not returned from the API if not Spot
		priority := string(containerservice.ScaleSetPriorityRegular)
		if props.ScaleSetPriority != "" {
			priority = string(props.ScaleSetPriority)
		}
		d.Set("priority", priority)

		proximityPlacementGroupId := ""
		if props.ProximityPlacementGroupID != nil {
			proximityPlacementGroupId = *props.ProximityPlacementGroupID
		}
		d.Set("proximity_placement_group_id", proximityPlacementGroupId)

		spotMaxPrice := -1.0
		if props.SpotMaxPrice != nil {
			spotMaxPrice = *props.SpotMaxPrice
		}
		d.Set("spot_max_price", spotMaxPrice)

		if err := d.Set("upgrade_settings", flattenUpgradeSettings(props.UpgradeSettings)); err != nil {
			return fmt.Errorf("setting `upgrade_settings`: %+v", err)
		}

		d.Set("vnet_subnet_id", props.VnetSubnetID)
		d.Set("vm_size", props.VMSize)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}
