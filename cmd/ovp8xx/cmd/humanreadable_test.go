package cmd

import (
	"strings"
	"testing"
)

func TestWriteHumanReadable(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{
			name: "nested object keeps the order of the keys",
			in:   `{"zeta":"OVP810","alpha":{"firmware":"1.2.3","build":42},"ok":true,"none":null}`,
			want: "zeta: OVP810\nalpha:\n  firmware: 1.2.3\n  build: 42\nok: true\nnone: null\n",
		},
		{
			name: "arrays",
			in:   `{"ports":["port0","port1"],"apps":[{"name":"a","id":1},{"name":"b"}]}`,
			want: "ports:\n  - port0\n  - port1\napps:\n  -\n    name: a\n    id: 1\n  -\n    name: b\n",
		},
		{
			name: "empty containers",
			in:   `{"a":{},"b":[],"c":""}`,
			want: "a: {}\nb: []\nc: \n",
		},
		{
			name: "numbers are not reformatted",
			in:   `{"serial":12345678901234567890,"temp":1.50}`,
			want: "serial: 12345678901234567890\ntemp: 1.50\n",
		},
		{name: "scalar", in: `"text"`, want: "text\n"},
		{name: "top level array", in: `[1,2]`, want: "- 1\n- 2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			if err := writeHumanReadable(&out, []byte(tt.in)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.String() != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", out.String(), tt.want)
			}
		})
	}
}

func TestWriteHumanReadableMalformed(t *testing.T) {
	for _, in := range []string{``, `{"a":`, `{"a":[1,`} {
		var out strings.Builder
		if err := writeHumanReadable(&out, []byte(in)); err == nil {
			t.Errorf("expected an error for %q, got %q", in, out.String())
		}
	}
}
