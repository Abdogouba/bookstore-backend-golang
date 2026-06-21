package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Maximum allowed image size (5 MB).
const MaxImageSize = 5 << 20 // 5 * 1024 * 1024 bytes

// Allowed image extensions.
var AllowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// SaveUploadedImage saves an uploaded image to disk
// and returns its relative path.
//
// If no image is uploaded, it returns an empty path
// and no error.
func SaveUploadedImage(
	c *gin.Context,
	formField string,
) (string, error) {

	// -------------------------
	// Get uploaded file
	// -------------------------

	file, err := c.FormFile(formField)

	// No image uploaded.
	if err != nil {
		return "", nil
	}

	// -------------------------
	// Validate file size
	// -------------------------

	if file.Size > MaxImageSize {
		return "", fmt.Errorf("image size must not exceed 5 MB")
	}

	// -------------------------
	// Validate file extension
	// -------------------------

	extension := strings.ToLower(
		filepath.Ext(file.Filename),
	)

	if !AllowedImageExtensions[extension] {
		return "", fmt.Errorf("unsupported image format")
	}

	// -------------------------
	// Create upload directory
	// -------------------------

	uploadDir := "uploads/books"

	err = os.MkdirAll(
		uploadDir,
		os.ModePerm,
	)

	if err != nil {
		return "", err
	}

	// -------------------------
	// Generate unique filename
	// -------------------------

	fileName := fmt.Sprintf(
		"%d%s",
		time.Now().UnixNano(),
		extension,
	)

	fullPath := filepath.Join(
		uploadDir,
		fileName,
	)

	// -------------------------
	// Save image
	// -------------------------

	err = c.SaveUploadedFile(
		file,
		fullPath,
	)

	if err != nil {
		return "", err
	}

	return fullPath, nil
}