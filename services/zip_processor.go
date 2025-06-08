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

	"mcphub/models"
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
	// Load zip archive in-memory
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %v", err)
	}

	// Prepare clean extraction directory
	extractDir := filepath.Join("extracted", strings.TrimSuffix(zipFileName, ".zip"))
	if err := os.RemoveAll(extractDir); err != nil {
		return nil, fmt.Errorf("failed to clean extract directory: %v", err)
	}
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create extract directory: %v", err)
	}

	// Unzip contents
	if err := zp.extractZip(reader, extractDir); err != nil {
		return nil, fmt.Errorf("failed to extract zip: %v", err)
	}

	// Locate and load mcp.json
	mcpConfig, mcpDir, err := zp.findAndParseMCPConfigFromDir(extractDir)
	if err != nil {
		return nil, err
	}

	// Generate Dockerfile from config
	dockerfileContent := zp.dockerfileGenerator.Generate(mcpConfig)

	// Save Dockerfile alongside mcp.json
	dockerfilePath := filepath.Join(mcpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile: %v", err)
	}

	// Build image using Dockerfile
	imageName := strings.ToLower(mcpConfig.Name)
	if err := zp.buildDockerImage(mcpDir, imageName); err != nil {
		return nil, fmt.Errorf("failed to build Docker image: %v", err)
	}

	// Save image as .tar archive
	tarFileName := strings.TrimSuffix(zipFileName, ".zip") + ".tar"
	tarFilePath := filepath.Join("output", tarFileName)
	if err := os.MkdirAll(filepath.Dir(tarFilePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}
	if err := zp.saveDockerImage(imageName, tarFilePath); err != nil {
		return nil, fmt.Errorf("failed to save Docker image: %v", err)
	}

	// Resolve absolute paths for response
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

// Extracts files from zip, optionally flattens single-folder zips
func (zp *ZipProcessor) extractZip(reader *zip.Reader, extractDir string) error {
	var commonPrefix string
	fileCount := 0

	// Detect common prefix (e.g., folder/) for flattening
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		fileCount++
		if fileCount == 1 {
			parts := strings.Split(file.Name, "/")
			if len(parts) > 1 {
				commonPrefix = parts[0] + "/"
			}
		} else if !strings.HasPrefix(file.Name, commonPrefix) {
			commonPrefix = "" // Mixed structure, don't flatten
			break
		}
	}

	// Extract all files to disk
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		targetPath := file.Name
		if commonPrefix != "" {
			targetPath = strings.TrimPrefix(file.Name, commonPrefix)
		}
		filePath := filepath.Join(extractDir, targetPath)

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		destFile, err := os.Create(filePath)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(destFile, rc)
		destFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Finds and validates mcp.json; prioritizes root-most if duplicates found
func (zp *ZipProcessor) findAndParseMCPConfigFromDir(extractDir string) (*models.MCPConfig, string, error) {
	var mcpFilePath string

	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "mcp.json" {
			if mcpFilePath == "" {
				mcpFilePath = path
			} else {
				// Prefer file closer to root
				rel1, _ := filepath.Rel(extractDir, mcpFilePath)
				rel2, _ := filepath.Rel(extractDir, path)
				if len(strings.Split(rel2, string(filepath.Separator))) < len(strings.Split(rel1, string(filepath.Separator))) {
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
		return nil, "", fmt.Errorf("mcp.json not found")
	}

	content, err := os.ReadFile(mcpFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read mcp.json: %v", err)
	}

	var mcpConfig models.MCPConfig
	if err := json.Unmarshal(content, &mcpConfig); err != nil {
		return nil, "", fmt.Errorf("failed to parse mcp.json: %v", err)
	}
	if mcpConfig.Name == "" || mcpConfig.Run.Command == "" {
		return nil, "", fmt.Errorf("mcp.json missing required fields")
	}

	return &mcpConfig, filepath.Dir(mcpFilePath), nil
}

// Builds Docker image from directory with Dockerfile
func (zp *ZipProcessor) buildDockerImage(buildContext, imageName string) error {
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Dir = buildContext
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}

// Saves image to tarball
func (zp *ZipProcessor) saveDockerImage(imageName, tarFilePath string) error {
	cmd := exec.Command("docker", "save", "-o", tarFilePath, imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker save failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}
