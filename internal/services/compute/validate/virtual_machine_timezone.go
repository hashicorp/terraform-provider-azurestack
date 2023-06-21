// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validate

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

// VirtualMachineTimeZone returns a case-sensitive validation function for the Time Zones for a Virtual Machine
func VirtualMachineTimeZone() pluginsdk.SchemaValidateFunc {
	return virtualMachineTimeZone(false)
}

// VirtualMachineTimeZone returns a case-insensitive validation function for the Time Zones for a Virtual Machine
func VirtualMachineTimeZoneCaseInsensitive() pluginsdk.SchemaValidateFunc {
	return virtualMachineTimeZone(true)
}

func virtualMachineTimeZone(ignoreCase bool) pluginsdk.SchemaValidateFunc {
	// Candidates are listed here: http://jackstromberg.com/2017/01/list-of-time-zones-consumed-by-azure/
	candidates := []string{
		"",
		"Afghanistan Standard Time",
		"Alaskan Standard Time",
		"Arab Standard Time",
		"Arabian Standard Time",
		"Arabic Standard Time",
		"Argentina Standard Time",
		"Atlantic Standard Time",
		"AUS Central Standard Time",
		"AUS Eastern Standard Time",
		"Azerbaijan Standard Time",
		"Azores Standard Time",
		"Bahia Standard Time",
		"Bangladesh Standard Time",
		"Belarus Standard Time",
		"Canada Central Standard Time",
		"Cape Verde Standard Time",
		"Caucasus Standard Time",
		"Cen. Australia Standard Time",
		"Central America Standard Time",
		"Central Asia Standard Time",
		"Central Brazilian Standard Time",
		"Central Europe Standard Time",
		"Central European Standard Time",
		"Central Pacific Standard Time",
		"Central Standard Time (Mexico)",
		"Central Standard Time",
		"China Standard Time",
		"Dateline Standard Time",
		"E. Africa Standard Time",
		"E. Australia Standard Time",
		"E. Europe Standard Time",
		"E. South America Standard Time",
		"Eastern Standard Time (Mexico)",
		"Eastern Standard Time",
		"Egypt Standard Time",
		"Ekaterinburg Standard Time",
		"Fiji Standard Time",
		"FLE Standard Time",
		"Georgian Standard Time",
		"GMT Standard Time",
		"Greenland Standard Time",
		"Greenwich Standard Time",
		"GTB Standard Time",
		"Hawaiian Standard Time",
		"India Standard Time",
		"Iran Standard Time",
		"Israel Standard Time",
		"Jordan Standard Time",
		"Kaliningrad Standard Time",
		"Korea Standard Time",
		"Libya Standard Time",
		"Line Islands Standard Time",
		"Magadan Standard Time",
		"Mauritius Standard Time",
		"Middle East Standard Time",
		"Montevideo Standard Time",
		"Morocco Standard Time",
		"Mountain Standard Time (Mexico)",
		"Mountain Standard Time",
		"Myanmar Standard Time",
		"N. Central Asia Standard Time",
		"Namibia Standard Time",
		"Nepal Standard Time",
		"New Zealand Standard Time",
		"Newfoundland Standard Time",
		"North Asia East Standard Time",
		"North Asia Standard Time",
		"Pacific SA Standard Time",
		"Pacific Standard Time (Mexico)",
		"Pacific Standard Time",
		"Pakistan Standard Time",
		"Paraguay Standard Time",
		"Romance Standard Time",
		"Russia Time Zone 10",
		"Russia Time Zone 11",
		"Russia Time Zone 3",
		"Russian Standard Time",
		"SA Eastern Standard Time",
		"SA Pacific Standard Time",
		"SA Western Standard Time",
		"Samoa Standard Time",
		"SE Asia Standard Time",
		"Singapore Standard Time",
		"South Africa Standard Time",
		"Sri Lanka Standard Time",
		"Syria Standard Time",
		"Taipei Standard Time",
		"Tasmania Standard Time",
		"Tokyo Standard Time",
		"Tonga Standard Time",
		"Turkey Standard Time",
		"Ulaanbaatar Standard Time",
		"US Eastern Standard Time",
		"US Mountain Standard Time",
		"UTC",
		"UTC+12",
		"UTC-02",
		"UTC-11",
		"Venezuela Standard Time",
		"Vladivostok Standard Time",
		"W. Australia Standard Time",
		"W. Central Africa Standard Time",
		"W. Europe Standard Time",
		"West Asia Standard Time",
		"West Pacific Standard Time",
		"Yakutsk Standard Time",
	}
	return validation.StringInSlice(candidates, ignoreCase)
}
