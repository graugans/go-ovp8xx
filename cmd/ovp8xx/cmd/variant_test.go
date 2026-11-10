package cmd

import "testing"

func envMap(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDetectVariant(t *testing.T) {
	tests := []struct {
		argv0 string
		env   map[string]string
		want  string
	}{
		{"ovp8xx", nil, "ovp8xx"},
		{"/usr/local/bin/ovp9xx", nil, "ovp9xx"},
		{"./o3cxxx", nil, "o3cxxx"},
		{`C:\tools\O3Cxxx.exe`, nil, "o3cxxx"},
		{"ovp8xx_linux_amd64", nil, "ovp8xx"},
		{"/tmp/go-build123/exe/ovp8xx", map[string]string{variantEnvVar: "ovp9xx"}, "ovp9xx"},
		{"o3cxxx", map[string]string{variantEnvVar: "ovp9xx"}, "o3cxxx"},
		{"/tmp/go-build123/exe/main", map[string]string{variantEnvVar: "OVP9XX"}, "ovp9xx"},
		{"main", map[string]string{variantEnvVar: "unknown"}, "ovp8xx"},
	}
	for _, tt := range tests {
		got := detectVariant(tt.argv0, envMap(tt.env))
		if got.name != tt.want {
			t.Errorf("detectVariant(%q, %v) = %q, want %q", tt.argv0, tt.env, got.name, tt.want)
		}
	}
}

func TestDefaultHost(t *testing.T) {
	v, _ := lookupVariant("o3cxxx")
	if got := v.defaultHost(envMap(nil)); got != v.defaultIP {
		t.Errorf("expected default IP, got %q", got)
	}
	if got := v.defaultHost(envMap(map[string]string{"OVP8XX_IP": "10.0.0.1"})); got != "10.0.0.1" {
		t.Errorf("expected OVP8XX_IP fallback, got %q", got)
	}
	got := v.defaultHost(envMap(map[string]string{"OVP8XX_IP": "10.0.0.1", "O3CXXX_IP": "10.0.0.2"}))
	if got != "10.0.0.2" {
		t.Errorf("expected O3CXXX_IP to take precedence, got %q", got)
	}
}
