package cmd

import (
	"fmt"
	"time"

	"github.com/graugans/go-ovp8xx/v2/pkg/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

func newMCPCommand() *cobra.Command {
	var timeout time.Duration
	var allowWrite bool
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Serve device tools over MCP stdio, read-only unless --allow-write is given",
		Long: `Serve the tools of one device over MCP stdio.

The server is read-only by default. With --allow-write it additionally exposes
set_config (temporary configuration changes) and save_init (persist them).
Reboot, factory reset and firmware updates are never exposed.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if timeout <= 0 {
				return fmt.Errorf("timeout must be positive")
			}
			host, err := cmd.Flags().GetString("ip")
			if err != nil {
				return err
			}
			var opts []mcpserver.Option
			if allowWrite {
				opts = append(opts, mcpserver.WithWrite())
			}
			return mcpserver.New(host, timeout, opts...).Run(cmd.Context(), &mcp.StdioTransport{})
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Maximum duration of each device tool or resource read")
	cmd.Flags().BoolVar(&allowWrite, "allow-write", false, "Expose the configuration write tools set_config and save_init")
	return cmd
}

func init() {
	rootCmd.AddCommand(newMCPCommand())
}
