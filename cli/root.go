package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global flag variables
var (
	yesFlag  bool
	detached bool
	portFlag string
	nameFlag string
)

var rootCmd = &cobra.Command{
	Use:   "mcphub",
	Short: "MCPHub CLI - Build and manage MCP servers",
	Long: `MCPHub CLI allows you to build and manage Model Context Protocol (MCP) servers.

Commands:
  init  - Initialize a new mcp.json configuration file
  push  - Build Docker image from MCP server zip file  
  pull  - Load Docker image from tar file
  run   - Run Docker container from loaded image`,
}

// Execute is the entry point for the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Register subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(runCmd)

	// Flags for 'init' command
	initCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Use default values without prompting")

	// Flags for 'run' command
	runCmd.Flags().BoolVarP(&detached, "detach", "d", true, "Run container in detached mode")
	runCmd.Flags().StringVarP(&portFlag, "port", "p", "", "Port mapping (e.g., 8080:8080)")
	runCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Container name (defaults to image name)")
}
