package validate

import (
	"regexp"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
)

func RepoRootFolder() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^\/(.*\/?)*$`),
		"Root folder must start with '/' and needs to be a valid git path")
}
