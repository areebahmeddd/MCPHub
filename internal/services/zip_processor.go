package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

func (zp *ZipProcessor) ProcessZip(zipData []byte, zipFileName string) (*models.DockerfileResponse, error) {
	// Create a zip reader
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %v", err)
	}

	// Extract zip file to a temporary directory
	extractDir := filepath.Join("extracted", strings.TrimSuffix(zipFileName, ".zip"))
	if err := os.RemoveAll(extractDir); err != nil {
		return nil, fmt.Errorf("failed to clean extract directory: %v", err)
	}
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create extract directory: %v", err)
	}
	// Extract all files
	if err := zp.extractZip(reader, extractDir); err != nil {
		return nil, fmt.Errorf("failed to extract zip: %v", err)
	}

	// Find and parse mcp.json
	mcpConfig, mcpDir, err := zp.findAndParseMCPConfigFromDir(extractDir)
	if err != nil {
		return nil, err
	}

	// Generate Dockerfile content
	dockerfileContent := zp.dockerfileGenerator.Generate(mcpConfig)

	// Write Dockerfile to the same directory as mcp.json (where the actual project files are)
	dockerfilePath := filepath.Join(mcpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile: %v", err)
	}
	// Build Docker image from the directory containing mcp.json and Dockerfile
	imageName := strings.ToLower(mcpConfig.Name)
	if err := zp.buildDockerImage(mcpDir, imageName); err != nil {
		return nil, fmt.Errorf("failed to build Docker image: %v", err)
	}

	// Save Docker image as tar
	tarFileName := strings.TrimSuffix(zipFileName, ".zip") + ".tar"
	tarFilePath := filepath.Join("output", tarFileName)
	if err := os.MkdirAll(filepath.Dir(tarFilePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	if err := zp.saveDockerImage(imageName, tarFilePath); err != nil {
		return nil, fmt.Errorf("failed to save Docker image: %v", err)
	}

	// Get absolute paths
	absExtractDir, _ := filepath.Abs(extractDir)
	absDockerfilePath, _ := filepath.Abs(dockerfilePath)
	absTarFilePath, _ := filepath.Abs(tarFilePath)

	return &models.DockerfileResponse{
		ExtractedPath:  absExtractDir,
		DockerfilePath: absDockerfilePath,
		ImageName:      imageName,
		TarFilePath:    absTarFilePath,
		Config:         *mcpConfig,
		Success:        true,
		Message:        fmt.Sprintf("Successfully processed %s. Docker image saved as %s", zipFileName, tarFileName),
	}, nil
}

// extractZip extracts all files from a zip reader to the specified directory
// Handles nested folder structures by flattening when there's a single top-level directory
func (zp *ZipProcessor) extractZip(reader *zip.Reader, extractDir string) error {
	// First, check if all files are nested under a single directory
	var commonPrefix string
	fileCount := 0
	
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		fileCount++
		
		if fileCount == 1 {
			// For the first file, set the common prefix
			parts := strings.Split(file.Name, "/")
			if len(parts) > 1 {
				commonPrefix = parts[0] + "/"
			}
		} else {
			// Check if this file also starts with the common prefix
			if commonPrefix != "" && !strings.HasPrefix(file.Name, commonPrefix) {
				commonPrefix = "" // No common prefix
				break
			}
		}
	}
	
	// Extract files, optionally removing the common prefix
	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Determine the target file path
		var targetPath string
		if commonPrefix != "" && strings.HasPrefix(file.Name, commonPrefix) {
			// Remove the common prefix to flatten the structure
			targetPath = strings.TrimPrefix(file.Name, commonPrefix)
		} else {
			targetPath = file.Name
		}
		
		// Create the full file path
		filePath := filepath.Join(extractDir, targetPath)

		// Create the directory for the file
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// Open the file in the zip
		rc, err := file.Open()
		if err != nil {
			return err
		}

		// Create the destination file
		destFile, err := os.Create(filePath)
		if err != nil {
			rc.Close()
			return err
		}

		// Copy the file content
		_, err = io.Copy(destFile, rc)
		destFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// findAndParseMCPConfigFromDir finds and parses mcp.json from the extracted directory
// Returns the config and the directory where mcp.json was found
func (zp *ZipProcessor) findAndParseMCPConfigFromDir(extractDir string) (*models.MCPConfig, string, error) {
	var mcpFilePath string

	// Walk through the directory to find mcp.json
	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "mcp.json" {
			if mcpFilePath == "" {
				mcpFilePath = path
			} else {
				// If multiple mcp.json files exist, prefer the one closer to root
				relPath1, _ := filepath.Rel(extractDir, mcpFilePath)
				relPath2, _ := filepath.Rel(extractDir, path)
				if len(strings.Split(relPath2, string(filepath.Separator))) < len(strings.Split(relPath1, string(filepath.Separator))) {
					mcpFilePath = path
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, "", fmt.Errorf("error walking directory: %v", err)
	}

	if mcpFilePath == "" {
		return nil, "", fmt.Errorf("mcp.json not found in the extracted directory")
	}

	// Read and parse the mcp.json file
	content, err := os.ReadFile(mcpFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read mcp.json: %v", err)
	}

	var mcpConfig models.MCPConfig
	if err := json.Unmarshal(content, &mcpConfig); err != nil {
		return nil, "", fmt.Errorf("failed to parse mcp.json: %v", err)
	}

	// Validate required fields
	if mcpConfig.Name == "" {
		return nil, "", fmt.Errorf("mcp.json is missing required 'name' field")
	}
	if mcpConfig.Run.Command == "" {
		return nil, "", fmt.Errorf("mcp.json is missing required 'run.command' field")
	}

	// Return the config and the directory containing mcp.json
	mcpDir := filepath.Dir(mcpFilePath)
	return &mcpConfig, mcpDir, nil
}

// buildDockerImage builds a Docker image from the extracted directory
func (zp *ZipProcessor) buildDockerImage(buildContext, imageName string) error {
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Dir = buildContext
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %v\nOutput: %s", err, string(output))
	}
	
	return nil
}

// saveDockerImage saves a Docker image as a tar file
func (zp *ZipProcessor) saveDockerImage(imageName, tarFilePath string) error {
	cmd := exec.Command("docker", "save", "-o", tarFilePath, imageName)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker save failed: %v\nOutput: %s", err, string(output))
	}
	
	return nil
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
