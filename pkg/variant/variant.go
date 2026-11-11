// Package variant describes the device families sharing the OVP8xx API
// (OVP8xx, OVP9xx, O3Cxxx) and resolves their defaults, e.g. the device host
// taken from environment variables.
//
// The package does not access the process environment on its own. Callers pass
// the program name and a getenv function, e.g. os.Getenv.
package variant

import "strings"

// Variant describes a device family. All device families share the same API,
// they only differ in naming and defaults.
type Variant struct {
	// Name selects the variant, e.g. "ovp9xx". It is also the binary name of
	// the multicall CLI.
	Name string
	// DisplayName is used in help texts, e.g. "OVP9xx".
	DisplayName string
	// DefaultIP is used when no host is given and no environment variable is set.
	DefaultIP string
}

// EnvVar can be used to select the variant when the program name cannot be
// used, e.g. on Windows or with "go run".
const EnvVar = "OVP8XX_VARIANT"

// Default is the variant used when no other variant is selected.
var Default = Variant{Name: "ovp8xx", DisplayName: "OVP8xx", DefaultIP: "192.168.0.69"}

var variants = []Variant{
	Default,
	{Name: "ovp9xx", DisplayName: "OVP9xx", DefaultIP: "192.168.0.69"},
	{Name: "o3cxxx", DisplayName: "O3Cxxx", DefaultIP: "192.168.0.69"},
}

// All returns all known variants. The default variant is the first element.
func All() []Variant {
	return append([]Variant(nil), variants...)
}

// Lookup returns the variant with the given name, e.g. "ovp9xx".
func Lookup(name string) (Variant, bool) {
	for _, v := range variants {
		if v.Name == name {
			return v, true
		}
	}
	return Variant{}, false
}

// IPEnvVar returns the name of the environment variable holding the device IP, e.g. "OVP9XX_IP".
func (v Variant) IPEnvVar() string {
	return strings.ToUpper(v.Name) + "_IP"
}

// IPEnvHint returns the environment variables consulted by DefaultHost in
// order of precedence, e.g. "OVP9XX_IP or OVP8XX_IP". It is meant for help texts.
func (v Variant) IPEnvHint() string {
	hint := v.IPEnvVar()
	if v != Default {
		hint += " or " + Default.IPEnvVar()
	}
	return hint
}

// DefaultHost returns the host to use when none is given explicitly.
// The variant specific environment variable takes precedence over OVP8XX_IP,
// which is kept as a fallback for backward compatibility. If neither is set,
// DefaultIP is returned.
func (v Variant) DefaultHost(getenv func(string) string) string {
	if ip := getenv(v.IPEnvVar()); ip != "" {
		return ip
	}
	if ip := getenv(Default.IPEnvVar()); ip != "" {
		return ip
	}
	return v.DefaultIP
}

// Detect selects the variant busybox style by the name the program was
// invoked with (argv0). A program name of a non default variant (e.g. "ovp9xx")
// always wins. Otherwise the OVP8XX_VARIANT environment variable is consulted,
// and if that is unset or unknown the default variant (ovp8xx) is used.
func Detect(argv0 string, getenv func(string) string) Variant {
	// Split on both separators so the detection behaves the same on all platforms.
	base := argv0[strings.LastIndexAny(argv0, `/\`)+1:]
	base = strings.ToLower(base)
	base = strings.TrimSuffix(base, ".exe")
	if v, ok := Lookup(base); ok && v != Default {
		return v
	}
	if v, ok := Lookup(strings.ToLower(getenv(EnvVar))); ok {
		return v
	}
	return Default
}
