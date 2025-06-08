package services

import (
	"testing"

	"mcphub/models"

	"github.com/stretchr/testify/assert"
)

func TestDockerfileGenerator_Generate(t *testing.T) {
	generator := NewDockerfileGenerator()

	t.Run("Python application", func(t *testing.T) {
		config := models.MCPConfig{
			Name:        "python-app",
			Version:     "1.0.0",
			Description: "Python application",
			Run: models.RunConfig{
				Command: "python3",
				Args:    []string{"app.py"},
				Port:    5000,
			},
		}

		output := generator.Generate(&config)

		assert.Contains(t, output, "FROM python:3.11-slim")
		assert.Contains(t, output, "EXPOSE 5000")
		assert.Contains(t, output, `CMD ["python3", "app.py"]`)
		assert.Contains(t, output, "requirements.txt")
	})

	t.Run("Node.js application", func(t *testing.T) {
		config := models.MCPConfig{
			Name:        "node-app",
			Version:     "1.0.0",
			Description: "Node.js application",
			Run: models.RunConfig{
				Command: "node",
				Args:    []string{"server.js"},
				Port:    8080,
			},
		}

		output := generator.Generate(&config)

		assert.Contains(t, output, "FROM node:18-alpine")
		assert.Contains(t, output, "EXPOSE 8080")
		assert.Contains(t, output, `CMD ["node", "server.js"]`)
		assert.Contains(t, output, "npm install")
	})
}
