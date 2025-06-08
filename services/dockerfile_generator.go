package services

import (
	"fmt"
	"strings"

	"mcphub/models"
)

type DockerfileGenerator struct{}

func NewDockerfileGenerator() *DockerfileGenerator {
	return &DockerfileGenerator{}
}

func (dg *DockerfileGenerator) Generate(config *models.MCPConfig) string {
	var dockerfile strings.Builder

	// Determine base image based on command
	baseImage := dg.getBaseImage(config.Run.Command)
	dockerfile.WriteString(fmt.Sprintf("FROM %s\n\n", baseImage))

	// Set working directory
	dockerfile.WriteString("WORKDIR /app\n\n")

	// Add metadata
	dockerfile.WriteString(fmt.Sprintf("LABEL name=\"%s\"\n", config.Name))
	dockerfile.WriteString(fmt.Sprintf("LABEL version=\"%s\"\n", config.Version))
	dockerfile.WriteString(fmt.Sprintf("LABEL description=\"%s\"\n", config.Description))
	if config.Author != "" {
		dockerfile.WriteString(fmt.Sprintf("LABEL author=\"%s\"\n", config.Author))
	}
	dockerfile.WriteString("\n")

	// Copy application files
	dockerfile.WriteString("COPY . .\n\n")

	// Install dependencies based on project type
	dg.addInstallCommands(&dockerfile, config.Run.Command)

	// Expose port if specified
	if config.Run.Port > 0 {
		dockerfile.WriteString(fmt.Sprintf("EXPOSE %d\n\n", config.Run.Port))
	}
	// Add health check if port is specified
	if config.Run.Port > 0 {
		dockerfile.WriteString("HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \\\n")
		dockerfile.WriteString(fmt.Sprintf("  CMD curl -f http://localhost:%d/health || exit 1\n\n", config.Run.Port))
	}

	// Set the command
	cmdArgs := append([]string{config.Run.Command}, config.Run.Args...)
	dockerfile.WriteString(fmt.Sprintf("CMD %s\n", dg.formatCommand(cmdArgs)))

	return dockerfile.String()
}

func (dg *DockerfileGenerator) getBaseImage(command string) string {
	switch command {
	case "node":
		return "node:18-alpine"
	case "python", "python3":
		return "python:3.11-slim"
	case "go":
		return "golang:1.21-alpine"
	default:
		return "ubuntu:22.04"
	}
}

func (dg *DockerfileGenerator) addInstallCommands(dockerfile *strings.Builder, command string) {
	switch command {
	case "node":
		dockerfile.WriteString("# Install dependencies\n")
		dockerfile.WriteString("RUN if [ -f package.json ]; then npm install --only=production; fi\n")
		dockerfile.WriteString("RUN if [ -f yarn.lock ]; then yarn install --production; fi\n\n")
	case "python", "python3":
		dockerfile.WriteString("# Install dependencies\n")
		dockerfile.WriteString("RUN if [ -f requirements.txt ]; then pip install --no-cache-dir -r requirements.txt; fi\n")
		dockerfile.WriteString("RUN if [ -f pyproject.toml ]; then pip install uv && uv pip install --system .; fi\n")
		dockerfile.WriteString("RUN if [ -f Pipfile ]; then pip install pipenv && pipenv install --system --deploy; fi\n\n")
	case "go":
		dockerfile.WriteString("# Install dependencies\n")
		dockerfile.WriteString("RUN if [ -f go.mod ]; then go mod download; fi\n")
		dockerfile.WriteString("RUN if [ -f go.mod ]; then go build -o main .; fi\n\n")
	default:
		dockerfile.WriteString("# Add any custom installation commands here\n\n")
	}
}

func (dg *DockerfileGenerator) formatCommand(cmdArgs []string) string {
	if len(cmdArgs) == 0 {
		return "[\"\"]"
	}

	// Format as JSON array for CMD instruction
	var quotedArgs []string
	for _, arg := range cmdArgs {
		quotedArgs = append(quotedArgs, fmt.Sprintf("\"%s\"", arg))
	}
	
	return fmt.Sprintf("[%s]", strings.Join(quotedArgs, ", "))
}
