package validate

import (
	"regexp"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
)

func SnapshotScheduleName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[^&%#/]{1,90}$`), `Data share snapshot schedule name should have length of 1 - 90, and cannot contain &%#/`,
	)
}
