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

// ProcessZip accepts zip data and filename, extracts contents, generates Dockerfile, builds and saves the image.
func (zp *ZipProcessor) ProcessZip(zipData []byte, zipFileName string) (*models.DockerfileResponse, error) {
	// Load zip archive from byte slice
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %w", err)
	}

	// Prepare extraction directory, cleaning if it exists
	extractDir := filepath.Join("extracted", strings.TrimSuffix(zipFileName, ".zip"))
	if err := os.RemoveAll(extractDir); err != nil {
		return nil, fmt.Errorf("failed to clean extraction directory: %w", err)
	}
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// Extract all files from zip to extraction directory
	if err := zp.extractZip(reader, extractDir); err != nil {
		return nil, fmt.Errorf("failed to extract zip contents: %w", err)
	}

	// Locate and parse mcp.json file (configuration)
	mcpConfig, mcpDir, err := zp.findAndParseMCPConfigFromDir(extractDir)
	if err != nil {
		return nil, err
	}

	// Generate Dockerfile text from config
	dockerfileContent := zp.dockerfileGenerator.Generate(mcpConfig)

	// Write Dockerfile next to mcp.json
	dockerfilePath := filepath.Join(mcpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Build Docker image
	imageName := strings.ToLower(mcpConfig.Name)
	if err := zp.buildDockerImage(mcpDir, imageName); err != nil {
		return nil, err
	}

	// Save Docker image as tar archive in output directory
	tarFileName := strings.TrimSuffix(zipFileName, ".zip") + ".tar"
	tarFilePath := filepath.Join("output", tarFileName)
	if err := os.MkdirAll(filepath.Dir(tarFilePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	if err := zp.saveDockerImage(imageName, tarFilePath); err != nil {
		return nil, err
	}

	// Return absolute paths in response
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

// extractZip extracts files from the zip archive, flattening single-folder archives
func (zp *ZipProcessor) extractZip(reader *zip.Reader, extractDir string) error {
	var commonPrefix string
	fileCount := 0

	// Detect common prefix folder if all files share it
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
			commonPrefix = "" // Mixed structure, no flattening
			break
		}
	}

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

// findAndParseMCPConfigFromDir searches for the mcp.json file and parses it, preferring the shallowest one if multiple
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
				// Prefer mcp.json closest to root (shallower path)
				relOld, _ := filepath.Rel(extractDir, mcpFilePath)
				relNew, _ := filepath.Rel(extractDir, path)
				if len(strings.Split(relNew, string(filepath.Separator))) < len(strings.Split(relOld, string(filepath.Separator))) {
					mcpFilePath = path
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("error walking extracted directory: %w", err)
	}

	if mcpFilePath == "" {
		return nil, "", fmt.Errorf("mcp.json not found")
	}

	content, err := os.ReadFile(mcpFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read mcp.json: %w", err)
	}

	var mcpConfig models.MCPConfig
	if err := json.Unmarshal(content, &mcpConfig); err != nil {
		return nil, "", fmt.Errorf("failed to parse mcp.json: %w", err)
	}

	if mcpConfig.Name == "" || mcpConfig.Run.Command == "" {
		return nil, "", fmt.Errorf("mcp.json missing required fields 'name' or 'run.command'")
	}

	return &mcpConfig, filepath.Dir(mcpFilePath), nil
}

// buildDockerImage builds a Docker image from the given context directory with the specified image name
func (zp *ZipProcessor) buildDockerImage(buildContext, imageName string) error {
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Dir = buildContext
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// saveDockerImage saves the specified Docker image to a tarball
func (zp *ZipProcessor) saveDockerImage(imageName, tarFilePath string) error {
	cmd := exec.Command("docker", "save", "-o", tarFilePath, imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker save failed: %w\nOutput: %s", err, output)
	}
	return nil
}
