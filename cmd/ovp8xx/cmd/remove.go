/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/v2/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func removeCommand(cmd *cobra.Command, args []string) error {
	p := args[0]
	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(host),
	)

	return o3r.Remove(p)
}

// removeCmd represents the get command
var removeCmd = &cobra.Command{
	Use:   "remove <pointer>",
	Short: "remove the OVP8xx",
	RunE:  removeCommand,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires exactly one positional argument")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
