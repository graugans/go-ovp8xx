/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func diagnosticGetFilteredCommand(cmd *cobra.Command, args []string) error {

	filter, err := cmd.Flags().GetString("filter")
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

	diag := o3r.GetDiagnosticClient()
	result, err := diag.GetFiltered(
		*ovp8xx.NewConfig(
			ovp8xx.WitJSONString(filter),
		),
	)
	if err != nil {
		return err
	} else {
		fmt.Printf("%s\n", result)
	}
	return nil
}

func diagnosticGetFilterSchemaCommand(cmd *cobra.Command, args []string) error {

	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(host),
	)

	diag := o3r.GetDiagnosticClient()
	result, err := diag.GetFilterSchema()
	if err != nil {
		return err
	} else {
		fmt.Printf("%s\n", result)
	}
	return nil
}

var diagnosticCmd = &cobra.Command{
	Use:   "diagnostic",
	Short: "Interact with diagnostic subsystem of the OVP8xx",
}

var diagnosticGetFilteredCmd = &cobra.Command{
	Use:   "getFiltered",
	Short: "Retrieve the filtered diagnostic JSON from the device",
	RunE:  diagnosticGetFilteredCommand,
}

var diagnosticGetFilterSchemaCmd = &cobra.Command{
	Use:   "getFilterSchema",
	Short: "Retrieve the diagnostic filter schema JSON from the device",
	RunE:  diagnosticGetFilterSchemaCommand,
}

func init() {
	rootCmd.AddCommand(diagnosticCmd)
	diagnosticCmd.AddCommand(diagnosticGetFilteredCmd)
	diagnosticCmd.AddCommand(diagnosticGetFilterSchemaCmd)
	diagnosticGetFilteredCmd.Flags().String("filter", "{\"state\": \"active\"}", "A JSON filter representation")
}
