package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	Use:   "pull <image_name>",
	Short: "Import a local Docker image from tar file",
	Long:  "Load a Docker image from a tar file that was created with mcphub push",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !dockerAvailable() {
			fmt.Println("❌ Docker is not running or not installed. Please start Docker and try again.")
			return
		}

		imageName := args[0]
		var tarFile string

		if strings.HasSuffix(imageName, ".tar") {
			tarFile = imageName
		} else {
			outputPath := filepath.Join("output", imageName+".tar")
			currentPath := imageName + ".tar"

			if _, err := os.Stat(outputPath); err == nil {
				tarFile = outputPath
			} else if _, err := os.Stat(currentPath); err == nil {
				tarFile = currentPath
			} else {
				fmt.Printf("❌ Could not find tar file for image '%s'\n", imageName)
				fmt.Printf("   Looked for: %s or %s\n", outputPath, currentPath)
				return
			}
		}

		fmt.Printf("🐳 Loading Docker image from %s...\n", tarFile)

		loadCmd := exec.Command("docker", "load", "-i", tarFile)
		loadOut, err := loadCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ Failed to load image: %s\n", string(loadOut))
			return
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
	},
}
