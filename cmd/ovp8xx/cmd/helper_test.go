package cmd

import (
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

func TestJSONPointers(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		args  []string
		want  []string
	}{
		{"none", nil, nil, []string{}},
		{"arguments", nil, []string{"/a", "/b"}, []string{"/a", "/b"}},
		{"flags", []string{"-p", "/a", "--pointer", "/b"}, nil, []string{"/a", "/b"}},
		{"mixed", []string{"-p", "/a"}, []string{"/b"}, []string{"/a", "/b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			addPointerFlag(cmd, pointerFlagUsage)
			if err := cmd.ParseFlags(tt.flags); err != nil {
				t.Fatal(err)
			}
			got, err := jsonPointers(cmd, tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if got == nil || !slices.Equal(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
