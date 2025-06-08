package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	client *s3.Client
	bucket string
}

func NewS3Service() (*S3Service, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	client := s3.NewFromConfig(cfg)
	return &S3Service{
		client: client,
		bucket: "mcp-servers", // You can make this configurable
	}, nil
}

// PushMCP uploads a tar file to S3
func (s *S3Service) PushMCP(author, imageName, tarPath string) error {
	objectKey := fmt.Sprintf("%s/%s.tar", author, imageName)

	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("error opening tar file: %v", err)
	}
	defer file.Close()

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String("application/gzip"),
	})

	if err != nil {
		return fmt.Errorf("error uploading to S3: %v", err)
	}

	// Clean up the local tar file after successful upload
	if err := os.Remove(tarPath); err != nil {
		return fmt.Errorf("error removing local tar file: %v", err)
	}

	return nil
}

// PullMCP downloads a tar file from S3
func (s *S3Service) PullMCP(author, imageName string) error {
	objectKey := fmt.Sprintf("%s/%s.tar", author, imageName)

	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("error downloading from S3: %v", err)
	}
	defer result.Body.Close()

	// Create downloaded directory if it doesn't exist
	downloadedDir := "downloaded"
	if err := os.MkdirAll(downloadedDir, 0755); err != nil {
		return fmt.Errorf("error creating downloaded directory: %v", err)
	}

	// Create the output file
	outputPath := filepath.Join(downloadedDir, fmt.Sprintf("%s.tar", imageName))
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	// Copy the S3 object body to the file
	if _, err := io.Copy(file, result.Body); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

// ListMCPs lists all MCPs in the S3 bucket
func (s *S3Service) ListMCPs() ([]string, error) {
	result, err := s.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects: %v", err)
	}

	var mcps []string
	for _, obj := range result.Contents {
		key := *obj.Key
		if strings.HasSuffix(key, ".tar") {
			// Remove .tar extension and add to list
			mcps = append(mcps, strings.TrimSuffix(key, ".tar"))
		}
	}

	return mcps, nil
} 