package cli

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"mcphub/services"

	"github.com/spf13/cobra"
)

// Check if Docker is available
func dockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

var pullCmd = &cobra.Command{
	Use:   "pull <author/image-name>",
	Short: "Download and import a Docker image from S3",
	Long:  "Download a Docker image from S3 and load it into Docker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !dockerAvailable() {
			return fmt.Errorf("❌ Docker is not running or not installed. Please start Docker and try again")
		}

		// Parse author/image-name format
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format. Use: author/image-name")
		}
		author := parts[0]
		imageName := parts[1]

		// Initialize S3 service
		s3Service, err := services.NewS3Service()
		if err != nil {
			return fmt.Errorf("failed to initialize S3 service: %v", err)
		}

		// Download from S3
		if err := s3Service.PullMCP(author, imageName); err != nil {
			return fmt.Errorf("failed to download from S3: %v", err)
		}

		// Load the Docker image
		tarFile := filepath.Join("downloaded", imageName+".tar")
		fmt.Printf("🐳 Loading Docker image from %s...\n", tarFile)

		loadCmd := exec.Command("docker", "load", "-i", tarFile)
		loadOut, err := loadCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to load image: %s", string(loadOut))
		}

		output := strings.TrimSpace(string(loadOut))
		fmt.Println("✅ Image loaded successfully!")
		if output != "" {
			fmt.Printf("📝 Docker output: %s\n", output)
		}

		// Attempt to extract the loaded image name from the output
		if strings.Contains(output, "Loaded image:") {
			parts := strings.Split(output, "Loaded image:")
			if len(parts) > 1 {
				loadedImage := strings.TrimSpace(parts[1])
				fmt.Printf("🏷️  Image: %s\n", loadedImage)
				fmt.Printf("💡 You can now run: mcphub run %s\n", strings.Split(loadedImage, ":")[0])
			}
		}

		return nil
	},
}
