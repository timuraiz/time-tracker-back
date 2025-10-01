package handlers

import (
	"log"
	"net/http"
	"time-tracker/database"
	"time-tracker/models"
	"time-tracker/supabase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateProfile creates a new profile for the authenticated user
func CreateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Check if profile already exists
	var existingProfile models.Profile
	if err := database.DB.Where("id = ?", uid).First(&existingProfile).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Profile already exists"})
		return
	}

	// Parse request body
	var req models.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create profile
	profile := models.Profile{
		ID:   uid,
		Name: req.Name,
	}

	if err := database.DB.Create(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
		return
	}

	c.JSON(http.StatusCreated, profile)
}

// GetProfile retrieves the current user's profile
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var profile models.Profile
	if err := database.DB.Where("id = ?", userID).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	response := models.UserResponse{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		ProfilePictureURL: profile.ProfilePictureURL,
	}

	c.JSON(http.StatusOK, response)
}

// UploadProfilePicture handles profile picture upload
func UploadProfilePicture(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	log.Println("HERE lala 3")

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get file from form
	log.Println("HERE 6")

	file, header, err := c.Request.FormFile("file")

	log.Println("HERE 7")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Get Supabase client
	supabaseClient := supabase.GetClient()

	// Get profile
	var profile models.Profile
	if err := database.DB.Where("id = ?", uid).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	// Delete old profile picture if exists
	if profile.ProfilePictureURL != nil && *profile.ProfilePictureURL != "" {
		if err := supabaseClient.DeleteProfilePicture(*profile.ProfilePictureURL); err != nil {
			// Log error but don't fail the upload
			log.Printf("Failed to delete old profile picture: %v\n", err)
		}
	}

	// Get user token from context
	userToken, exists := c.Get("access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User token not found"})
		return
	}

	token, ok := userToken.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token format"})
		return
	}

	// Upload new profile picture
	publicURL, err := supabaseClient.UploadProfilePicture(uid, file, header, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update profile record with new profile picture URL
	if err := database.DB.Model(&profile).Update("profile_picture_url", publicURL).Error; err != nil {
		// If database update fails, try to delete the uploaded file
		supabaseClient.DeleteProfilePicture(publicURL)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile picture"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Profile picture uploaded successfully",
		"profile_picture_url": publicURL,
	})
}

// DeleteProfilePicture handles profile picture deletion
func DeleteProfilePicture(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get profile
	var profile models.Profile
	if err := database.DB.Where("id = ?", uid).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	// Check if profile has a profile picture
	if profile.ProfilePictureURL == nil || *profile.ProfilePictureURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No profile picture to delete"})
		return
	}

	// Get Supabase client
	supabaseClient := supabase.GetClient()

	// Delete from storage
	if err := supabaseClient.DeleteProfilePicture(*profile.ProfilePictureURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile picture from storage"})
		return
	}

	// Update database to remove URL
	if err := database.DB.Model(&profile).Update("profile_picture_url", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile picture deleted successfully"})
}