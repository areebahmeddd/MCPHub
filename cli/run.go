package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <image_name>",
	Short: "Run a Docker container from a loaded image",
	Long:  `Start a Docker container from an image that was loaded with mcphub pull`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !dockerAvailable() {
			fmt.Println("❌ Docker is not running or not installed. Please start Docker and try again.")
			return
		}

		imageName := args[0]
		containerName := nameFlag
		if containerName == "" {
			containerName = imageName
		}

		// Build docker run command
		dockerArgs := []string{"run"}

		if detached {
			dockerArgs = append(dockerArgs, "-d")
		} else {
			dockerArgs = append(dockerArgs, "-it")
		}

		dockerArgs = append(dockerArgs, "--name", containerName)

		if portFlag != "" {
			dockerArgs = append(dockerArgs, "-p", portFlag)
		}

		dockerArgs = append(dockerArgs, imageName)

		fmt.Printf("🚀 Running container from image '%s'...\n", imageName)
		if detached {
			fmt.Printf("🔧 Command: docker %s\n", strings.Join(dockerArgs, " "))
		}

		dockerCmd := exec.Command("docker", dockerArgs...)

		if detached {
			runOut, err := dockerCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("❌ Failed to run container: %s\n", string(runOut))
				return
			}
			containerID := strings.TrimSpace(string(runOut))
			fmt.Println("✅ Container started successfully!")
			fmt.Printf("🆔 Container ID: %s\n", containerID)
			fmt.Printf("📋 Container Name: %s\n", containerName)
			if portFlag != "" {
				fmt.Printf("🌐 Port mapping: %s\n", portFlag)
			}
			fmt.Printf("💡 To view logs: docker logs %s\n", containerName)
			fmt.Printf("💡 To stop: docker stop %s\n", containerName)
		} else {
			// Run container interactively in foreground
			dockerCmd.Stdout = os.Stdout
			dockerCmd.Stderr = os.Stderr
			dockerCmd.Stdin = os.Stdin

			if err := dockerCmd.Run(); err != nil {
				fmt.Printf("❌ Container exited with error: %v\n", err)
			}
		}
	},
}
