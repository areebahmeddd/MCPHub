package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"mcphub/internal/models"
)

func TestZipProcessor_ProcessZip(t *testing.T) {
	// Create a test zip file with mcp.json
	mcpConfig := models.MCPConfig{
		Name:        "test-app",
		Version:     "1.0.0",
		Description: "Test application",
		Author:      "Test Author",
		License:     "MIT",
		Keywords:    []string{"test", "mcp"},
		Repository: models.Repository{
			Type: "git",
			URL:  "https://github.com/test/test-app",
		},
		Run: models.RunConfig{
			Command: "node",
			Args:    []string{"index.js"},
			Port:    3000,
		},
	}

	// Create zip file in memory
	zipBuffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(zipBuffer)

	// Add mcp.json to zip
	mcpJSON, _ := json.Marshal(mcpConfig)
	mcpFile, _ := zipWriter.Create("mcp.json")
	mcpFile.Write(mcpJSON)

	// Add a dummy index.js file
	indexFile, _ := zipWriter.Create("index.js")
	indexFile.Write([]byte("console.log('Hello World');"))

	zipWriter.Close()

	// Test the processor
	processor := NewZipProcessor()
	result, err := processor.ProcessZip(zipBuffer.Bytes())

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "test-app", result.Config.Name)
	assert.Contains(t, result.Dockerfile, "FROM node:18-alpine")
	assert.Contains(t, result.Dockerfile, "EXPOSE 3000")
	assert.Contains(t, result.Dockerfile, "CMD [\"node\", \"index.js\"]")
}

func TestDockerfileGenerator_Generate(t *testing.T) {
	generator := NewDockerfileGenerator()

	// Test Node.js application
	nodeConfig := models.MCPConfig{
		Name:        "node-app",
		Version:     "1.0.0",
		Description: "Node.js application",
		Run: models.RunConfig{
			Command: "node",
			Args:    []string{"server.js"},
			Port:    8080,
		},
	}

	dockerfile := generator.Generate(&nodeConfig)
	assert.Contains(t, dockerfile, "FROM node:18-alpine")
	assert.Contains(t, dockerfile, "EXPOSE 8080")
	assert.Contains(t, dockerfile, "CMD [\"node\", \"server.js\"]")
	assert.Contains(t, dockerfile, "npm install")

	// Test Python application
	pythonConfig := models.MCPConfig{
		Name:        "python-app",
		Version:     "1.0.0",
		Description: "Python application",
		Run: models.RunConfig{
			Command: "python3",
			Args:    []string{"app.py"},
			Port:    5000,
		},
	}

	dockerfile2 := generator.Generate(&pythonConfig)
	assert.Contains(t, dockerfile2, "FROM python:3.11-slim")
	assert.Contains(t, dockerfile2, "EXPOSE 5000")
	assert.Contains(t, dockerfile2, "CMD [\"python3\", \"app.py\"]")
	assert.Contains(t, dockerfile2, "requirements.txt")
}
