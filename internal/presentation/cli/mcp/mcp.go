package mcp

import (
	"context"
	"fmt"
	"os"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/logger"
	jmcp "github.com/SamyRai/juleson/internal/mcp"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"github.com/spf13/cobra"
)

func NewCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run the Juleson MCP server",
		Long:  "Run the Juleson MCP server over stdio for Jules session and developer workflow tools.",
	}

	var version bool
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve MCP over stdio",
		Long:  "Serve the Juleson MCP server over stdin/stdout. Diagnostics are written to stderr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if version {
				info := core.GetVersionInfo()
				fmt.Print(core.FormatVersion(info))
				return nil
			}
			// Redirect logger to stderr to avoid corrupting MCP JSON-RPC over stdout
			logger.SetupGlobalWithOutput(cfg.Jules.DebugLog, os.Stderr)
			return jmcp.RunStdio(context.Background(), cfg)
		},
	}
	serveCmd.Flags().BoolVar(&version, "version", false, "Print version and exit without starting the MCP server")
	cmd.AddCommand(serveCmd)

	return cmd
}
