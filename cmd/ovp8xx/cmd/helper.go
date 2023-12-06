package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

type helperConfig struct {
	pretty   bool
	host     string
	pointers []string
}

func (c *helperConfig) printJSONResult(data string) error {
	var message string = data
	if c.pretty {
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

func (c *helperConfig) jsonPointers() []string {
	return c.pointers
}

func NewHelper(cmd *cobra.Command) (helperConfig, error) {
	var conf = helperConfig{}
	var err error
	conf.pretty, err = cmd.Flags().GetBool("pretty")
	if err != nil {
		return conf, err
	}
	conf.host, err = rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return conf, err
	}
	// Pointers can be empty
	conf.pointers, _ = cmd.Flags().GetStringSlice("pointer")
	return conf, nil
}
