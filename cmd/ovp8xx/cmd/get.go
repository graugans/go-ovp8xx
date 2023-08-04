/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func getCommand(cmd *cobra.Command, args []string) error {
	o3r := ovp8xx.NewClient()
	query := []string{"/device"}
	result, err := o3r.Get(query)
	if err != nil {
		return err
	} else {
		fmt.Printf("%s\n", result)
	}
	return nil
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve the JSON configuration from the device",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: getCommand,
}

func init() {
	rootCmd.AddCommand(getCmd)
}
