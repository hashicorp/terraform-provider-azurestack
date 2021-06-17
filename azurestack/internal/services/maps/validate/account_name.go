package validate

import (
	"regexp"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
)

func AccountName() pluginsdk.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9]{1}[A-Za-z0-9._-]{1,}$`),
		"First character must be alphanumeric. Subsequent character(s) must be any combination of alphanumeric, underscore (_), period (.), or hyphen (-).")
}
