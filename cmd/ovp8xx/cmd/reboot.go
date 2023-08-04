/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// rebootCmd represents the reboot command
var rebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot the device",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("reboot called")
	},
}

func init() {
	rootCmd.AddCommand(rebootCmd)
}
