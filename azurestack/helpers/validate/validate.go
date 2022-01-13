package validate

import (
	"fmt"
	"net"
)

func PortNumber(i interface{}, k string) (warnings []string, errors []error) {
	return validatePortNumber(i, k, false)
}

func validatePortNumber(i interface{}, k string, allowZero bool) (warnings []string, errors []error) {
	v, ok := i.(int)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be int", k))
		return
	}

	if allowZero && v == 0 {
		return
	}

	if v < 1 || 65535 < v {
		errors = append(errors, fmt.Errorf("%q is not a valid port number: %d", k, v))
	}

	return warnings, errors
}

func IPv4Address(i interface{}, k string) (warnings []string, errors []error) {
	return validateIpv4Address(i, k, false)
}

func IPv4AddressOrEmpty(i interface{}, k string) (warnings []string, errors []error) {
	return validateIpv4Address(i, k, true)
}

func validateIpv4Address(i interface{}, k string, allowEmpty bool) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if v == "" && allowEmpty {
		return
	}

	ip := net.ParseIP(v)
	if four := ip.To4(); four == nil {
		errors = append(errors, fmt.Errorf("%q is not a valid IPv4 address: %q", k, v))
	}

	return warnings, errors
}
