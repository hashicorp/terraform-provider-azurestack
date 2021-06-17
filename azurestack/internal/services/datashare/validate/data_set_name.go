package validate

import (
	"regexp"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
)

func DataSetName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[\w-]{2,90}$`), `Dataset name can only contain number, letters, - and _, and must be between 2 and 90 characters long.`,
	)
}
