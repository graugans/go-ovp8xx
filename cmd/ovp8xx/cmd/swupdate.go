/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/graugans/go-ovp8xx/pkg/swupdater"
	"github.com/spf13/cobra"
)

func swupdateCommand(cmd *cobra.Command, args []string) error {
	var err error
	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return fmt.Errorf("cannot get host: %w", err)
	}

	port, err := cmd.Flags().GetUint16("port")
	if err != nil {
		return fmt.Errorf("cannot get port: %w", err)
	}

	filename, err := cmd.Flags().GetString("file")
	if err != nil {
		return fmt.Errorf("cannot get filename: %w", err)
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return fmt.Errorf("cannot get timeout: %w", err)
	}

	fmt.Printf("Updating firmware on %s:%d with file %s (%v)\n", host, port, filename, timeout)

	swu := swupdater.NewSWUpdater(host, port)

	err = swu.Update(filename, timeout)
	if err != nil {
		return fmt.Errorf("software update failed: %w", err)
	}

	return nil
}

// swupdateCmd represents the swupdate command
var swupdateCmd = &cobra.Command{
	Use:   "swupdate",
	Short: "Update the firmware on the device",
	RunE:  swupdateCommand,
}

func init() {
	rootCmd.AddCommand(swupdateCmd)
	swupdateCmd.Flags().String("file", "", "A file conatining the firmware image")
	swupdateCmd.Flags().Uint16("port", 8080, "Port number for SWUpdate")
	swupdateCmd.Flags().Duration("timeout", 5*time.Minute, "The timeout for the upload")
}
