package supabase

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	ProfilePicturesBucket = "profile-pictures"
	MaxFileSize           = 5 * 1024 * 1024 // 5MB
)

// UploadProfilePicture uploads a profile picture to Supabase Storage
func (c *Client) UploadProfilePicture(userID uuid.UUID, file multipart.File, header *multipart.FileHeader, userToken string) (string, error) {
	// Validate file size
	if header.Size > MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of 5MB")
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		return "", fmt.Errorf("invalid file type: %s (allowed: jpg, jpeg, png, webp)", ext)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s/%s%s", userID.String(), uuid.New().String(), ext)

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Determine content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(fileBytes)
	}

	// Upload to Supabase Storage
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.URL, ProfilePicturesBucket, filename)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("apikey", c.AnonKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Return public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.URL, ProfilePicturesBucket, filename)
	return publicURL, nil
}

// DeleteProfilePicture deletes a profile picture from Supabase Storage
func (c *Client) DeleteProfilePicture(pictureURL string) error {
	// Extract filename from URL
	// Format: https://xxx.supabase.co/storage/v1/object/public/profile-pictures/filename
	parts := strings.Split(pictureURL, "/storage/v1/object/public/"+ProfilePicturesBucket+"/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid picture URL format")
	}
	filename := parts[1]

	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.URL, ProfilePicturesBucket, filename)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("apikey", c.ServiceKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}