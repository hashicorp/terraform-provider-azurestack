// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utils

import "net"

func NormalizeIPv6Address(ipv6 interface{}) string {
	if ipv6 == nil || ipv6.(string) == "" {
		return ""
	}
	r := net.ParseIP(ipv6.(string))
	if r == nil {
		return ""
	}
	return r.String()
}
