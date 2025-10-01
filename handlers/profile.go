package handlers

import (
	"log"
	"net/http"
	"time"
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

	// Get time entries for stats
	var entries []models.TimeEntry
	if err := database.DB.Where("user_id = ? AND duration > ?", userID, 0).Order("start_time DESC").Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch time entries"})
		return
	}

	// Calculate stats
	var totalHours float32
	for _, entry := range entries {
		totalHours += float32(entry.Duration)
	}
	totalSessions := len(entries)
	totalHours = totalHours / 60 / 60

	var dayilyAvg float32
	if totalSessions > 0 {
		dayilyAvg = totalHours / float32(totalSessions)
	}

	currentStreak := 0
	if len(entries) > 0 {
		// Track unique days
		dayMap := make(map[string]bool)
		for _, entry := range entries {
			day := entry.StartTime.Format("2006-01-02")
			dayMap[day] = true
		}

		// Convert to sorted slice of days
		days := make([]string, 0, len(dayMap))
		for day := range dayMap {
			days = append(days, day)
		}

		// Sort days in descending order
		for i := 0; i < len(days)-1; i++ {
			for j := i + 1; j < len(days); j++ {
				if days[i] < days[j] {
					days[i], days[j] = days[j], days[i]
				}
			}
		}

		// Count consecutive days
		if len(days) > 0 {
			currentStreak = 1
			for k := 1; k < len(days); k++ {
				prevDay, _ := time.Parse("2006-01-02", days[k-1])
				currDay, _ := time.Parse("2006-01-02", days[k])
				diff := prevDay.Sub(currDay).Hours() / 24

				if diff == 1 {
					currentStreak++
				} else {
					break
				}
			}
		}
	}

	// Calculate level based on total hours
	level, levelColor := getLevelByHours(totalHours)

	// Calculate rank among all users
	var rank int
	err := database.DB.Raw(`
		SELECT COUNT(*) + 1
		FROM (
			SELECT user_id, SUM(duration) as total_duration
			FROM time_entries
			WHERE duration > 0
			GROUP BY user_id
			HAVING SUM(duration) > (
				SELECT COALESCE(SUM(duration), 0)
				FROM time_entries
				WHERE user_id = ? AND duration > 0
			)
		) ranked_users
	`, userID).Scan(&rank).Error

	if err != nil {
		rank = 0
	}

	response := models.UserResponse{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		ProfilePictureURL: profile.ProfilePictureURL,
		TotalHours:        float32(int(totalHours*10+0.5)) / 10,
		TotalSessions:     totalSessions,
		CurrentStreak:     currentStreak,
		DayilyAvg:         float32(int(dayilyAvg*10+0.5)) / 10,
		Rank:              rank,
		Level:             level,
		LevelColor:        levelColor,
		CreatedAt:         profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// getLevelByHours returns a level and color based on total hours tracked
func getLevelByHours(hours float32) (string, string) {
	switch {
	case hours < 1:
		return "Newbie", "#00d4ff"
	case hours < 5:
		return "Beginner", "#00e5ff"
	case hours < 10:
		return "Learner", "#1de9b6"
	case hours < 25:
		return "Apprentice", "#00e676"
	case hours < 50:
		return "Practitioner", "#76ff03"
	case hours < 100:
		return "Skilled", "#aeea00"
	case hours < 150:
		return "Experienced", "#ffd600"
	case hours < 200:
		return "Professional", "#ffab00"
	case hours < 300:
		return "Expert", "#ff6d00"
	case hours < 400:
		return "Veteran", "#ff3d00"
	case hours < 500:
		return "Master", "#ff1744"
	case hours < 750:
		return "Grandmaster", "#f50057"
	case hours < 1000:
		return "Elite", "#d500f9"
	case hours < 1500:
		return "Champion", "#aa00ff"
	case hours < 2000:
		return "Hero", "#651fff"
	case hours < 3000:
		return "Legend", "#3d5afe"
	case hours < 4000:
		return "Mythic", "#2979ff"
	case hours < 5000:
		return "Immortal", "#00b0ff"
	case hours < 7500:
		return "Divine", "#00e5ff"
	default:
		return "Eternal", "#1de9b6"
	}
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

// GetLeaderboard returns top 5 users by total time tracked
func GetLeaderboard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUserID, _ := userID.(uuid.UUID)

	type LeaderboardEntry struct {
		UserID            uuid.UUID `json:"user_id"`
		Name              string    `json:"name"`
		ProfilePictureURL *string   `json:"profile_picture_url,omitempty"`
		TotalHours        float32   `json:"total_hours"`
		Level             string    `json:"level"`
		LevelColor        string    `json:"level_color"`
		Rank              int       `json:"rank"`
		CurrentStreak     int       `json:"current_streak"`
		IsCurrentUser     bool      `json:"is_current_user"`
	}

	var results []struct {
		UserID        uuid.UUID
		TotalDuration float32
	}

	err := database.DB.Raw(`
		SELECT user_id, SUM(duration) / 3600.0 as total_duration
		FROM time_entries
		WHERE duration > 0
		GROUP BY user_id
		ORDER BY total_duration DESC
		LIMIT 5
	`).Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	leaderboard := make([]LeaderboardEntry, 0, len(results))

	for i, result := range results {
		var user models.User
		if err := database.DB.Where("id = ?", result.UserID).First(&user).Error; err != nil {
			continue
		}

		var profile models.Profile
		database.DB.Where("id = ?", result.UserID).First(&profile)

		// Get time entries for streak calculation
		var entries []models.TimeEntry
		database.DB.Where("user_id = ? AND duration > ?", result.UserID, 0).Order("start_time DESC").Find(&entries)

		currentStreak := 0
		if len(entries) > 0 {
			// Track unique days
			dayMap := make(map[string]bool)
			for _, entry := range entries {
				day := entry.StartTime.Format("2006-01-02")
				dayMap[day] = true
			}

			// Convert to sorted slice of days
			days := make([]string, 0, len(dayMap))
			for day := range dayMap {
				days = append(days, day)
			}

			// Sort days in descending order
			for i := 0; i < len(days)-1; i++ {
				for j := i + 1; j < len(days); j++ {
					if days[i] < days[j] {
						days[i], days[j] = days[j], days[i]
					}
				}
			}

			// Count consecutive days
			if len(days) > 0 {
				currentStreak = 1
				for k := 1; k < len(days); k++ {
					prevDay, _ := time.Parse("2006-01-02", days[k-1])
					currDay, _ := time.Parse("2006-01-02", days[k])
					diff := prevDay.Sub(currDay).Hours() / 24

					if diff == 1 {
						currentStreak++
					} else {
						break
					}
				}
			}
		}

		level, levelColor := getLevelByHours(result.TotalDuration)
		totalHours := float32(int(result.TotalDuration*10+0.5)) / 10

		leaderboard = append(leaderboard, LeaderboardEntry{
			UserID:            user.ID,
			Name:              profile.Name,
			ProfilePictureURL: profile.ProfilePictureURL,
			TotalHours:        totalHours,
			Level:             level,
			LevelColor:        levelColor,
			Rank:              i + 1,
			CurrentStreak:     currentStreak,
			IsCurrentUser:     user.ID == currentUserID,
		})
	}

	c.JSON(http.StatusOK, leaderboard)
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

