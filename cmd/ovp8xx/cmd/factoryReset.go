/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

// factoryResetCmd represents the reset command
var factoryResetCmd = &cobra.Command{
	Use:   "factoryReset",
	Short: "Performs a factory reset of the device",
	Long: `Sometime one wants a fresh start.

The command factoryReset resets all settings to their defaults and erases any addtional data like Docker containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keepNetworkSettings, err := cmd.Flags().GetBool("keepnetworksettings")
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
		return o3r.FactoryReset(keepNetworkSettings)
	},
}

func init() {
	rootCmd.AddCommand(factoryResetCmd)
	factoryResetCmd.Flags().Bool("keepnetworksettings", true, "Weather to keep the network settings or reset them to the factory defaults")
}
