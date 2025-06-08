package config

import (
	"os"
)

type Config struct {
	Port        string
	UploadDir   string
	MaxFileSize int64
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	return &Config{
		Port:        port,
		UploadDir:   uploadDir,
		MaxFileSize: 100 * 1024 * 1024, // 100MB
	}
}
