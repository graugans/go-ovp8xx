/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func getInitCommand(cmd *cobra.Command, args []string) error {

	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(host),
	)

	if result, err := o3r.GetInit(); err != nil {
		return err
	} else {
		fmt.Printf("%s\n", result)
	}
	return nil
}

// getCmd represents the get command
var getInitCmd = &cobra.Command{
	Use:   "getInit",
	Short: "Retrieve the init JSON configuration from the device",
	Long: `The OVP8xx provides a way to store a configuration on the device

NOTE: This shall be used with care, because it may lead to an system which is no
longer useable when the expectation from the safed configuration is no longer met.`,
	RunE: getInitCommand,
}

func init() {
	rootCmd.AddCommand(getInitCmd)
}
