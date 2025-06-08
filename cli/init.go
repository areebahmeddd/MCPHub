package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mcphub/models"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new mcp.json file",
	Long:  "Create a new mcp.json configuration file in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		var mcp models.MCPConfig

		if yesFlag {
			cwd, err := os.Getwd()
			projectName := "my-project"

			if err == nil {
				projectName = filepath.Base(cwd)
			}

			mcp = models.MCPConfig{
				Name:        projectName,
				Version:     "1.0.0",
				Description: "",
				Author:      "",
				License:     "MIT",
				Keywords:    []string{},
				Repository: models.Repository{
					Type: "git",
					URL:  "",
				},
				Run: models.RunConfig{
					Command: "node",
					Args:    []string{"index.js"},
					Port:    5050,
				},
			}
		} else {
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Project name: ")
			mcp.Name = readLine(reader)

			fmt.Print("Version (1.0.0): ")
			version := readLine(reader)
			if version == "" {
				version = "1.0.0"
			}
			mcp.Version = version

			fmt.Print("Description: ")
			mcp.Description = readLine(reader)

			fmt.Print("Author: ")
			mcp.Author = readLine(reader)

			fmt.Print("License (MIT): ")
			license := readLine(reader)
			if license == "" {
				license = "MIT"
			}
			mcp.License = license

			fmt.Print("Keywords (comma separated): ")
			keywords := readLine(reader)
			if keywords != "" {
				mcp.Keywords = strings.Split(keywords, ",")
				for i, keyword := range mcp.Keywords {
					mcp.Keywords[i] = strings.TrimSpace(keyword)
				}
			} else {
				mcp.Keywords = []string{}
			}

			fmt.Print("Repository type (git): ")
			repoType := readLine(reader)
			if repoType == "" {
				repoType = "git"
			}
			mcp.Repository.Type = repoType

			fmt.Print("Repository URL: ")
			mcp.Repository.URL = readLine(reader)

			fmt.Print("Run command (node): ")
			command := readLine(reader)
			if command == "" {
				command = "node"
			}
			mcp.Run.Command = command

			fmt.Print("Run arguments (index.js): ")
			argsStr := readLine(reader)
			if argsStr == "" {
				mcp.Run.Args = []string{"index.js"}
			} else {
				mcp.Run.Args = strings.Split(argsStr, ",")
				for i, arg := range mcp.Run.Args {
					mcp.Run.Args[i] = strings.TrimSpace(arg)
				}
			}

			fmt.Print("Port (5050): ")
			var port int
			_, err := fmt.Scanf("%d\n", &port)
			if err != nil || port == 0 {
				port = 5050
			}
			mcp.Run.Port = port
		}

		file, err := os.Create("mcp.json")
		if err != nil {
			fmt.Printf("❌ Error creating mcp.json: %v\n", err)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")

		if err := encoder.Encode(mcp); err != nil {
			fmt.Printf("❌ Error writing mcp.json: %v\n", err)
			return
		}

		fmt.Println("✅ mcp.json created successfully!")
		fmt.Printf("📋 Project: %s v%s\n", mcp.Name, mcp.Version)

		if mcp.Description != "" {
			fmt.Printf("📝 Description: %s\n", mcp.Description)
		}
	},
}

func readLine(reader *bufio.Reader) string {
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}
