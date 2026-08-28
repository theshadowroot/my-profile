package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	Client *cloudinary.Cloudinary
}

func NewCloudinaryService() (*CloudinaryService, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" {
		return nil, fmt.Errorf("CLOUDINARY_CLOUD_NAME is not set")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("CLOUDINARY_API_KEY is not set")
	}

	if apiSecret == "" {
		return nil, fmt.Errorf("CLOUDINARY_API_SECRET is not set")
	}

	cld, err := cloudinary.NewFromParams(
		cloudName,
		apiKey,
		apiSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	return &CloudinaryService{
		Client: cld,
	}, nil

}

func (s *CloudinaryService) UploadImage(
	ctx context.Context,
	file multipart.File,
	filename string,
	folder string,
) (string, error) {
	if file == nil {
		return "", fmt.Errorf("image file is required")
	}

	result, err := s.Client.Upload.Upload(
		ctx,
		file,
		uploader.UploadParams{
			Folder:       folder,
			PublicID:     filename,
			ResourceType: "image",
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to Cloudinary: %w", err)
	}

	if result.SecureURL == "" {
		return "", fmt.Errorf("Cloudinary returned an empty image URL")
	}

	return result.SecureURL, nil

}

func (s *CloudinaryService) UploadPDF(
	ctx context.Context,
	file multipart.File,
	filename string,
	folder string,
) (string, error) {
	if file == nil {
		return "", fmt.Errorf("PDF file is required")
	}

	result, err := s.Client.Upload.Upload(
		ctx,
		file,
		uploader.UploadParams{
			Folder:       folder,
			PublicID:     filename,
			ResourceType: "raw",
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF to Cloudinary: %w", err)
	}

	if result.SecureURL == "" {
		return "", fmt.Errorf("Cloudinary returned an empty PDF URL")
	}

	return result.SecureURL, nil

}

func readUploadFile(file multipart.File) ([]byte, error) {
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset uploaded file: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset uploaded file: %w", err)
	}

	return data, nil

}
