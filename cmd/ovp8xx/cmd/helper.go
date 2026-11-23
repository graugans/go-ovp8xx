package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/graugans/go-ovp8xx/v2/pkg/pcic"
	"github.com/spf13/cobra"
)

type helperConfig struct {
	pretty bool
	host   string
	port   uint16
}

func (c *helperConfig) printJSONResult(data string) error {
	var message = data
	if c.prettyPrint() {
		var js json.RawMessage
		if err := json.Unmarshal([]byte(data), &js); err != nil {
			return errors.New("malformed json")
		}
		jsonIndented, err := json.MarshalIndent(js, "", "  ")
		if err != nil {
			return err
		}
		message = string(jsonIndented)
	}
	fmt.Print(message)
	return nil
}

func (c *helperConfig) hostname() string {
	return c.host
}

func (c *helperConfig) remotePort() uint16 {
	return c.port
}

// prettyPrint returns a boolean value indicating whether the output should be pretty-printed.
func (c *helperConfig) prettyPrint() bool {
	return c.pretty
}

// pointerFlagUsage is the usage text of the --pointer flag.
const pointerFlagUsage = "A JSON pointer to be queried, can be repeated; the same as passing it as an argument"

// addPointerFlag adds the --pointer/-p flag to cmd.
func addPointerFlag(cmd *cobra.Command, usage string) {
	cmd.Flags().StringSliceP("pointer", "p", nil, usage)
}

// jsonPointers returns the JSON pointers given by --pointer/-p followed by the
// ones given as arguments. Both forms can be mixed. The result is never nil,
// an empty list queries the complete document.
func jsonPointers(cmd *cobra.Command, args []string) ([]string, error) {
	pointers, err := cmd.Flags().GetStringSlice("pointer")
	if err != nil {
		return nil, fmt.Errorf("cannot get the JSON pointers: %w", err)
	}
	return append(append([]string{}, pointers...), args...), nil
}

func NewHelper(cmd *cobra.Command) (helperConfig, error) {
	var conf = helperConfig{}
	var err error
	conf.pretty, err = cmd.Flags().GetBool("pretty")
	if err != nil {
		// In case no pretty flag is set, we default to false
		conf.pretty = false
	}
	conf.host, err = rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return conf, err
	}

	// Port can be empty
	conf.port, err = cmd.Flags().GetUint16("port")
	if err != nil {
		conf.port = pcic.DefaultPort
	}

	return conf, nil
}
