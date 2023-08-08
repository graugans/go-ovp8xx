/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Change the configuration of the device",
	Long: `This will send a JSON configuration to the device
	
NOTE: Sending smaller bites of JSON will reduce the chewing costs on the device.
Reduce your JSON to the max, only send the bare minimum required to achieve your goals
Not just copy paste the output of a get all ("") with a small change into a set.
Better extract the object you want to change and only send this`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("invalid number of arguments provided to 'set'")
		}

		if !json.Valid([]byte(args[0])) {
			return fmt.Errorf("the import argument is not valid JSON")
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
				ovp8xx.WitJSONString(args[0]),
			),
		)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
