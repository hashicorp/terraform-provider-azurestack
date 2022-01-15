package sdk

import "github.com/hashicorp/terraform-provider-azurestack/internal/az/resourceid"

// SetID uses the specified ID Formatter to set the Resource ID
func (rmd ResourceMetaData) SetID(formatter resourceid.Formatter) {
	rmd.ResourceData.SetId(formatter.ID())
}
