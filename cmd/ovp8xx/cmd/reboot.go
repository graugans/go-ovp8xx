/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func rebootCommand(cmd *cobra.Command, args []string) error {
	var swu bool
	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(host),
	)

	if swu, err = cmd.Flags().GetBool("swupdate"); err != nil {
		return err
	}

	if swu {
		return o3r.RebootToSWUpdate()
	}

	return o3r.Reboot()
}

// rebootCmd represents the get command
var rebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot the OVP8xx",
	RunE:  rebootCommand,
}

func init() {
	rootCmd.AddCommand(rebootCmd)
	rebootCmd.Flags().Bool("swupdate", false, "Reboot to the SWUpdate mode")
}
