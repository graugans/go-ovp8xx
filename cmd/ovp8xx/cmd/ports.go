/*
Copyright © 2024 Christian Ege <ch@ege.io>

SPDX-License-Identifier: Apache-2.0
*/
package cmd

import (
	"fmt"

	"github.com/graugans/go-ovp8xx/v2/pkg/ovp8xx"
	"github.com/spf13/cobra"
)

func listPortsCommand(cmd *cobra.Command, args []string) error {

	host, err := rootCmd.PersistentFlags().GetString("ip")
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(host),
	)

	ports, err := o3r.Ports()
	if err != nil {
		return fmt.Errorf("unable to retrieve the port list from the device: %w", err)
	}
	for _, port := range ports {
		fmt.Printf("[%s] state: %s,\ttype: %s,\tPCIC Port: %d\n",
			port.Name,
			port.State,
			port.Type,
			port.PCIC.Port,
		)
	}

	return nil
}

var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Interact with ports, also known as heads, connected to the OVP8xx",
	Long: `The ports are labeled from Port0 to Port5 on the VPU itself. They are 
connected to the device through coaxial cables. Port6 is optional and is connected 
internally. Port6 represents the inertial measurement unit (IMU).
`,
}

var portsListCmd = &cobra.Command{
	Use:   "ls",
	Short: "Retrieve the list of connected ports from the device",
	RunE:  listPortsCommand,
}

func init() {
	rootCmd.AddCommand(portsCmd)
	portsCmd.AddCommand(portsListCmd)
}
