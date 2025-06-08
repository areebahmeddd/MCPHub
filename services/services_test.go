package services

import (
	"testing"

	"mcphub/models"

	"github.com/stretchr/testify/assert"
)

func TestDockerfileGenerator_Generate(t *testing.T) {
	generator := NewDockerfileGenerator()

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

	dockerfile1 := generator.Generate(&pythonConfig)
	assert.Contains(t, dockerfile1, "FROM python:3.11-slim")
	assert.Contains(t, dockerfile1, "EXPOSE 5000")
	assert.Contains(t, dockerfile1, "CMD [\"python3\", \"app.py\"]")
	assert.Contains(t, dockerfile1, "requirements.txt")

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

	dockerfile2 := generator.Generate(&nodeConfig)
	assert.Contains(t, dockerfile2, "FROM node:18-alpine")
	assert.Contains(t, dockerfile2, "EXPOSE 8080")
	assert.Contains(t, dockerfile2, "CMD [\"node\", \"server.js\"]")
	assert.Contains(t, dockerfile2, "npm install")
}
