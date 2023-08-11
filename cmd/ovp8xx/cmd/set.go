/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Change the configuration of the device",
	Long: `This will send a JSON configuration to the device

If no argument is provided the "set" command reads input data from stdin.

The input data is validated to contain a proper JSON string.
There is only one JSON string allowed as an input

NOTE: Sending smaller bites of JSON will reduce the chewing costs on the device.
Reduce your JSON to the max, only send the bare minimum required to achieve your goals
Not just copy paste the output of a get all ("") with a small change into a set.
Better extract the object you want to change and only send this`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		var data []byte
		var err error

		if len(args) == 1 {
			// A commandline arg was given let's use this
			data = []byte(args[0])
		} else {
			// No input Argument was provided try to ready from stdin
			if data, err = io.ReadAll(cmd.InOrStdin()); err != nil {
				return err
			}

		}
		if !json.Valid(data) {
			return fmt.Errorf("the input argument is not valid JSON")
		}

		host, err := rootCmd.PersistentFlags().GetString("ip")
		if err != nil {
			return err
		}

		o3r := ovp8xx.NewClient(
			ovp8xx.WithHost(host),
		)

		return o3r.Set(
			*ovp8xx.NewConfig(
				ovp8xx.WitJSONString(string(data)),
			),
		)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
