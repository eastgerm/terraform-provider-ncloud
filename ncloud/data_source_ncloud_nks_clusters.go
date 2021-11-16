package ncloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	RegisterDataSource("ncloud_nks_clusters", dataSourceNcloudNKSClusters())
}

func dataSourceNcloudNKSClusters() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNcloudNKSClustersRead,
		Schema: map[string]*schema.Schema{
			"cluster_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceNcloudNKSClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*ProviderConfig)
	if !config.SupportVPC {
		return diag.FromErr(NotSupportClassic("dataSource `ncloud_nks_clusters`"))
	}

	clusters, err := getNKSClusters(ctx, config)
	if err != nil {
		return diag.FromErr(err)
	}

	var cNames []*string
	for _, cluster := range clusters {
		cNames = append(cNames, cluster.Name)
	}

	d.SetId(config.RegionCode)
	d.Set("cluster_names", cNames)

	return nil
}
