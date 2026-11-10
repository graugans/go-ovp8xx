package cmd

import (
	"os"
	"strings"
)

// variant describes a device family supported by this multicall binary.
// All device families share the same API, they only differ in naming and defaults.
type variant struct {
	// name is the binary name which selects the variant, e.g. "ovp9xx"
	name string
	// displayName is used in help texts, e.g. "OVP9xx"
	displayName string
	// defaultIP is used when neither the --ip flag nor an environment variable is set
	defaultIP string
}

// ipEnvVar returns the name of the environment variable holding the device IP, e.g. "OVP9XX_IP".
func (v variant) ipEnvVar() string {
	return strings.ToUpper(v.name) + "_IP"
}

// defaultHost returns the host to use when no --ip flag is given.
// The variant specific environment variable takes precedence over OVP8XX_IP,
// which is kept as a fallback for backward compatibility.
func (v variant) defaultHost(getenv func(string) string) string {
	if ip := getenv(v.ipEnvVar()); ip != "" {
		return ip
	}
	if ip := getenv(defaultVariant.ipEnvVar()); ip != "" {
		return ip
	}
	return v.defaultIP
}

// variantEnvVar can be used to select the variant when creating a link is not
// possible or inconvenient, e.g. on Windows or with "go run".
const variantEnvVar = "OVP8XX_VARIANT"

var defaultVariant = variant{name: "ovp8xx", displayName: "OVP8xx", defaultIP: "192.168.0.69"}

var variants = []variant{
	defaultVariant,
	{name: "ovp9xx", displayName: "OVP9xx", defaultIP: "192.168.0.69"},
	{name: "o3cxxx", displayName: "O3Cxxx", defaultIP: "192.168.0.69"},
}

// detectVariant selects the variant busybox style by the name the binary was
// invoked with (argv[0]). A link name of a non default variant (e.g. "ovp9xx")
// always wins. Otherwise the OVP8XX_VARIANT environment variable is consulted,
// and if that is unset or unknown the default variant (ovp8xx) is used.
func detectVariant(argv0 string, getenv func(string) string) variant {
	// Split on both separators so the detection behaves the same on all platforms.
	base := argv0[strings.LastIndexAny(argv0, `/\`)+1:]
	base = strings.ToLower(base)
	base = strings.TrimSuffix(base, ".exe")
	if v, ok := lookupVariant(base); ok && v != defaultVariant {
		return v
	}
	if v, ok := lookupVariant(strings.ToLower(getenv(variantEnvVar))); ok {
		return v
	}
	return defaultVariant
}

func lookupVariant(name string) (variant, bool) {
	for _, v := range variants {
		if v.name == name {
			return v, true
		}
	}
	return variant{}, false
}

// currentVariant is the variant the binary has been invoked as.
var currentVariant = detectVariant(os.Args[0], os.Getenv)
