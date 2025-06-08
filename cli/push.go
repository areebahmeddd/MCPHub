package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mcphub/services"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [zip-file]",
	Short: "Process an MCP server zip file and build Docker image",
	Long: `Process an MCP server zip file by:
1. Extracting the zip file
2. Finding and parsing mcp.json configuration
3. Generating a Dockerfile
4. Building a Docker image
5. Saving the image as a tar file`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
	zipFilePath := args[0]

	// Check if file exists
	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		return fmt.Errorf("zip file does not exist: %s", zipFilePath)
	}

	// Read the zip file
	zipData, err := os.ReadFile(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to read zip file: %v", err)
	}

	// Check file size (100MB limit)
	if len(zipData) > 100*1024*1024 {
		return fmt.Errorf("file size exceeds 100MB limit")
	}

	// Get just the filename from the path
	zipFileName := filepath.Base(zipFilePath)

	fmt.Printf("📦 Processing %s...\n", zipFileName)

	// Process the zip file using the existing service
	processor := services.NewZipProcessor()
	result, err := processor.ProcessZip(zipData, zipFileName)
	if err != nil {
		return fmt.Errorf("failed to process zip file: %v", err)
	}

	// Display results
	fmt.Println("✅ Success!")
	fmt.Printf("📁 Extracted to: %s\n", result.ExtractedPath)
	fmt.Printf("🐳 Dockerfile: %s\n", result.DockerfilePath)
	fmt.Printf("🏷️  Image name: %s\n", result.ImageName)
	fmt.Printf("📦 Docker image saved to: %s\n", result.TarFilePath)
	fmt.Printf("📋 MCP Server: %s v%s\n", result.Config.Name, result.Config.Version)

	if result.Config.Description != "" {
		fmt.Printf("📝 Description: %s\n", result.Config.Description)
	}

	if result.Config.Author != "" {
		fmt.Printf("👤 Author: %s\n", result.Config.Author)
	}

	if result.Config.Version != "" {
		fmt.Printf("🏷️  Version: %s\n", result.Config.Version)
	}

	if len(result.Config.Keywords) > 0 {
		fmt.Printf("🏷️  Keywords: %s\n", strings.Join(result.Config.Keywords, ", "))
	}

	return nil
}
