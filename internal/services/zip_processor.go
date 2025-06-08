package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"mcphub/internal/models"
)

type ZipProcessor struct {
	dockerfileGenerator *DockerfileGenerator
}

func NewZipProcessor() *ZipProcessor {
	return &ZipProcessor{
		dockerfileGenerator: NewDockerfileGenerator(),
	}
}

func (zp *ZipProcessor) ProcessZip(zipData []byte) (*models.DockerfileResponse, error) {
	// Create a zip reader
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %v", err)
	}

	// Find and parse mcp.json
	mcpConfig, err := zp.findAndParseMCPConfig(reader)
	if err != nil {
		return nil, err
	}

	// Generate Dockerfile
	dockerfile := zp.dockerfileGenerator.Generate(mcpConfig)

	return &models.DockerfileResponse{
		Dockerfile: dockerfile,
		Config:     *mcpConfig,
		Success:    true,
	}, nil
}

func (zp *ZipProcessor) findAndParseMCPConfig(reader *zip.Reader) (*models.MCPConfig, error) {
	var mcpFile *zip.File

	// First, look for mcp.json in the root directory
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "mcp.json") && !strings.Contains(file.Name, "/") {
			mcpFile = file
			break
		}
	}

	// If not found in root, look for mcp.json in any subdirectory
	if mcpFile == nil {
		for _, file := range reader.File {
			if strings.HasSuffix(file.Name, "mcp.json") {
				// Count the number of directory separators to find the shallowest one
				if mcpFile == nil || strings.Count(file.Name, "/") < strings.Count(mcpFile.Name, "/") {
					mcpFile = file
				}
			}
		}
	}

	if mcpFile == nil {
		return nil, fmt.Errorf("mcp.json not found in the zip file")
	}

	// Open and read the mcp.json file
	rc, err := mcpFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open mcp.json: %v", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read mcp.json: %v", err)
	}

	// Parse JSON
	var mcpConfig models.MCPConfig
	if err := json.Unmarshal(content, &mcpConfig); err != nil {
		return nil, fmt.Errorf("failed to parse mcp.json: %v", err)
	}

	// Validate required fields
	if mcpConfig.Name == "" {
		return nil, fmt.Errorf("mcp.json is missing required 'name' field")
	}
	if mcpConfig.Run.Command == "" {
		return nil, fmt.Errorf("mcp.json is missing required 'run.command' field")
	}

	return &mcpConfig, nil
}
