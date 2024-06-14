/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"
	"path/filepath"
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

	connectionTimeout, err := cmd.Flags().GetDuration("online")
	if err != nil {
		return fmt.Errorf("cannot get timeout: %w", err)
	}

	fmt.Printf("Updating firmware on %s:%d with file %s (%v)\n",
		host,
		port,
		filepath.Base(filename),
		timeout,
	)

	// notifications is a channel used to receive SWUpdaterNotification events.
	// It has a buffer size of 10 to allow for asynchronous processing.
	notifications := make(chan swupdater.SWUpdaterNotification, 10)

	// Print the messages as they come
	go func() {
		for n := range notifications {
			if value, ok := n["swupdater"]; ok {
				fmt.Println(value)
			}
			if value, ok := n["text"]; ok && n["type"] == "message" {
				fmt.Println(value)
			}
		}
	}()

	// Create a new SWUpdater instance with the specified host, port, and notifications.
	swu := swupdater.NewSWUpdater(host, port, notifications)
	if err = swu.Update(filename,
		connectionTimeout,
		timeout,
	); err != nil {
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
	swupdateCmd.Flags().Duration("online", 2*time.Minute, "The time to wait for the device to become available")
}
