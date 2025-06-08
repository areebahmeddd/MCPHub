package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mcphub/internal/models"
	"mcphub/internal/services"
)

func GenerateDockerfile(c *gin.Context) {
	// Get uploaded file
	file, header, err := c.Request.FormFile("zip")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.DockerfileResponse{
			Success: false,
			Message: "No zip file provided",
		})
		return
	}
	defer file.Close()

	// Validate file size (100MB limit)
	if header.Size > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, models.DockerfileResponse{
			Success: false,
			Message: "File size exceeds 100MB limit",
		})
		return
	}

	// Read file content
	fileContent := make([]byte, header.Size)
	_, err = file.Read(fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.DockerfileResponse{
			Success: false,
			Message: "Failed to read uploaded file",
		})
		return
	}
	// Process the zip file
	processor := services.NewZipProcessor()
	result, err := processor.ProcessZip(fileContent, header.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.DockerfileResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
