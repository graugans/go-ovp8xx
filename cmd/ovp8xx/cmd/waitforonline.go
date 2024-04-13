/*
Copyright © 2023 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/graugans/go-ovp8xx/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func waitForOnlineCommand(cmd *cobra.Command, args []string) error {
	var ok = false
	var err error
	helper, err := NewHelper(cmd)
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(helper.hostname()),
	)

	timeout, err := cmd.Flags().GetUint("timeout")
	if err != nil {
		return err
	}

	if ok, err = o3r.IsAvailable(time.Duration(timeout) * time.Second); err != nil {
		return err
	}
	if ok {
		fmt.Printf("The device is available now\n")
	}
	return nil
}

// waitForOnlineCmd represents the wait command
var waitForOnlineCmd = &cobra.Command{
	Use:   "WaitForOnline",
	Short: "Wait until the device is accessible",
	Long: `This command is maybe useful after a reboot or power on. 
It can be used to wait until the device can handle requests`,
	RunE: waitForOnlineCommand,
}

func init() {
	rootCmd.AddCommand(waitForOnlineCmd)
	waitForOnlineCmd.Flags().Uint("timeout", 120, "Timeout in Seconds to wait until the device is accessible")
}
