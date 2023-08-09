/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

// saveInitCmd represents the reboot command
var saveInitCmd = &cobra.Command{
	Use:   "saveInit",
	Short: "Saves the init configuration on the device",
	Long: `To store the configuration persistant on the device the command saveInit can be used.

A safed configuration persists a reboot. This is best used in combination with the "set" command.

Please use this with care. The scope should be as narrow as posible, to prevent any conflicts.
In case no JSON Pointer is provided the complete configuration is saved`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pointers, err := cmd.Flags().GetStringSlice("pointer")
		if err != nil {
			return err
		}
		host, err := rootCmd.PersistentFlags().GetString("ip")
		if err != nil {
			return err
		}

		o3r := ovp8xx.NewClient(
			ovp8xx.WithHost(host),
		)
		return o3r.SaveInit(pointers)
	},
}

func init() {
	rootCmd.AddCommand(saveInitCmd)
	saveInitCmd.Flags().StringSliceP("pointer", "p", []string{""}, "A JSON pointer to be saved, this defines the scope")
}
