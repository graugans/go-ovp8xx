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
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Serve read-only device tools over MCP stdio",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if timeout <= 0 {
				return fmt.Errorf("timeout must be positive")
			}
			host, err := cmd.Flags().GetString("ip")
			if err != nil {
				return err
			}
			return mcpserver.New(host, timeout).Run(cmd.Context(), &mcp.StdioTransport{})
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Maximum duration of each device tool or resource read")
	return cmd
}

func init() {
	rootCmd.AddCommand(newMCPCommand())
}
