/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"github.com/graugans/go-ovp8xx/v2/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

// saveInitCmd represents the reboot command
var saveInitCmd = &cobra.Command{
	Use:   "saveInit [pointers...]",
	Short: "Saves the init configuration on the device",
	Long: `To store the configuration persistent on the device the command saveInit can be used.

A saved configuration persists a reboot. This is best used in combination with the "set" command.

Please use this with care. The scope should be as narrow as possible, to prevent any conflicts.
In case no JSON Pointer is provided the complete configuration is saved.
The pointers can be given as arguments or with --pointer/-p, both forms can be mixed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pointers, err := jsonPointers(cmd, args)
		if err != nil {
			return err
		}
		if len(pointers) == 0 {
			// Keep the previous behaviour: save the complete configuration
			pointers = []string{""}
		}
		host, err := rootCmd.PersistentFlags().GetString("ip")
		if err != nil {
			return err
		}

		o3r := ovp8xx.NewClient(
			ovp8xx.WithHost(host),
		)
		return o3r.SaveInit(cmd.Context(), pointers)
	},
}

func init() {
	rootCmd.AddCommand(saveInitCmd)
	addPointerFlag(saveInitCmd, "A JSON pointer to be saved, this defines the scope; the same as passing it as an argument")
}
