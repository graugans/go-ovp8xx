/*
Copyright © 2024 Christian Ege <ch@ege.io>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/graugans/go-ovp8xx/v2/pkg/ovp8xx"
	"github.com/graugans/go-ovp8xx/v2/pkg/swupdater"
	"github.com/spf13/cobra"
)

// The modes a device can be waited for.
const (
	waitModeAuto       = "auto"
	waitModeProductive = "productive"
	waitModeSWUpdate   = "swupdate"
)

// availabilityCheck blocks until the device is available or ctx is done.
type availabilityCheck func(ctx context.Context) (bool, error)

// waitForAny runs the given checks concurrently and returns the mode of the
// first one reporting the device as available. The remaining checks are
// cancelled. If no check succeeds, the error of the last one is returned.
func waitForAny(ctx context.Context, checks map[string]availabilityCheck) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		mode string
		ok   bool
		err  error
	}
	// The channel is buffered, so the checks can terminate after a return.
	results := make(chan result, len(checks))
	for mode, check := range checks {
		go func() {
			ok, err := check(ctx)
			results <- result{mode: mode, ok: ok, err: err}
		}()
	}

	err := fmt.Errorf("device is not available")
	for range checks {
		r := <-results
		if r.ok && r.err == nil {
			return r.mode, nil
		}
		if r.err != nil {
			err = r.err
		}
	}
	if len(checks) > 1 && ctx.Err() != nil {
		// Do not blame a single mode if none of them was reached in time
		err = fmt.Errorf("device is not available: %w", ctx.Err())
	}
	return "", err
}

func waitForOnlineCommand(cmd *cobra.Command, args []string) error {
	helper, err := NewHelper(cmd)
	if err != nil {
		return err
	}

	timeout, err := cmd.Flags().GetUint("timeout")
	if err != nil {
		return err
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}

	swupdatePort, err := cmd.Flags().GetUint16("swupdate-port")
	if err != nil {
		return err
	}

	o3r := ovp8xx.NewClient(
		ovp8xx.WithHost(helper.hostname()),
	)
	swu := swupdater.NewSWUpdater(helper.hostname(), swupdatePort, nil)

	checks := map[string]availabilityCheck{
		waitModeProductive: func(ctx context.Context) (bool, error) {
			return o3r.IsAvailable(ctx, ovp8xx.AndPortsAreOnline())
		},
		waitModeSWUpdate: swu.IsAvailable,
	}
	switch mode {
	case waitModeAuto:
	case waitModeProductive, waitModeSWUpdate:
		checks = map[string]availabilityCheck{mode: checks[mode]}
	default:
		return fmt.Errorf("invalid mode %q, expected one of: %s, %s, %s",
			mode, waitModeAuto, waitModeProductive, waitModeSWUpdate)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(timeout)*time.Second)
	defer cancel()
	reached, err := waitForAny(ctx, checks)
	if err != nil {
		return err
	}
	fmt.Printf("The device is available now (%s mode)\n", reached)
	return nil
}

// waitForOnlineCmd represents the wait command
var waitForOnlineCmd = &cobra.Command{
	Use:   "WaitForOnline",
	Short: "Wait until the device is accessible",
	Long: `This command is maybe useful after a reboot or power on.
It can be used to wait until the device can handle requests.

A device is either running in the productive mode, where it is configured
via XML-RPC, or in the SWUpdate mode, where it accepts a firmware update.
The mode to wait for is selected with --mode:

  auto        wait for either of both modes, whichever is reached first (default)
  productive  wait until the device and its ports are initialized
  swupdate    wait until the SWUpdate service is reachable

The mode which was reached is part of the output.`,
	Args: cobra.NoArgs,
	RunE: waitForOnlineCommand,
}

func init() {
	rootCmd.AddCommand(waitForOnlineCmd)
	waitForOnlineCmd.Flags().Uint("timeout", 120, "Timeout in Seconds to wait until the device is accessible")
	waitForOnlineCmd.Flags().String("mode", waitModeAuto,
		fmt.Sprintf("The mode to wait for: %s, %s or %s", waitModeAuto, waitModeProductive, waitModeSWUpdate))
	waitForOnlineCmd.Flags().Uint16("swupdate-port", 8080, "Port number of the SWUpdate service")
}
