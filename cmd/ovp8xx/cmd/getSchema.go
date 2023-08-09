/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func getSchemaCommand(cmd *cobra.Command, args []string) error {

	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(host),
	)

	if result, err := o3r.GetSchema(); err != nil {
		return err
	} else {
		fmt.Printf("%s\n", result)
	}
	return nil
}

// getCmd represents the get command
var getSchemaCmd = &cobra.Command{
	Use:   "getSchema",
	Short: "Retrieve the currently used JSON schema from the device",
	RunE:  getSchemaCommand,
}

func init() {
	rootCmd.AddCommand(getSchemaCmd)
}
