package variant

import "testing"

func envMap(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDetect(t *testing.T) {
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
		{"/tmp/go-build123/exe/ovp8xx", map[string]string{EnvVar: "ovp9xx"}, "ovp9xx"},
		{"o3cxxx", map[string]string{EnvVar: "ovp9xx"}, "o3cxxx"},
		{"/tmp/go-build123/exe/main", map[string]string{EnvVar: "OVP9XX"}, "ovp9xx"},
		{"main", map[string]string{EnvVar: "unknown"}, "ovp8xx"},
	}
	for _, tt := range tests {
		got := Detect(tt.argv0, envMap(tt.env))
		if got.Name != tt.want {
			t.Errorf("Detect(%q, %v) = %q, want %q", tt.argv0, tt.env, got.Name, tt.want)
		}
	}
}

func TestDefaultHost(t *testing.T) {
	v, _ := Lookup("o3cxxx")
	if got := v.DefaultHost(envMap(nil)); got != v.DefaultIP {
		t.Errorf("expected default IP, got %q", got)
	}
	if got := v.DefaultHost(envMap(map[string]string{"OVP8XX_IP": "10.0.0.1"})); got != "10.0.0.1" {
		t.Errorf("expected OVP8XX_IP fallback, got %q", got)
	}
	got := v.DefaultHost(envMap(map[string]string{"OVP8XX_IP": "10.0.0.1", "O3CXXX_IP": "10.0.0.2"}))
	if got != "10.0.0.2" {
		t.Errorf("expected O3CXXX_IP to take precedence, got %q", got)
	}
}

func TestIPEnvHint(t *testing.T) {
	if got := Default.IPEnvHint(); got != "OVP8XX_IP" {
		t.Errorf("default variant: got %q", got)
	}
	v, _ := Lookup("ovp9xx")
	if got := v.IPEnvHint(); got != "OVP9XX_IP or OVP8XX_IP" {
		t.Errorf("ovp9xx: got %q", got)
	}
}

func TestAll(t *testing.T) {
	all := All()
	if len(all) != 3 || all[0] != Default {
		t.Fatalf("unexpected variants: %v", all)
	}
	all[0].Name = "changed"
	if All()[0] != Default {
		t.Error("All must return a copy")
	}
}
