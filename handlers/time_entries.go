package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"time-tracker/database"
	"time-tracker/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateTimeEntry(c *gin.Context) {
	var req models.TimeEntryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	timeEntry := models.TimeEntry{
		UserID:      userID.(uuid.UUID),
		ProjectID:   req.ProjectID,
		StartTime:   time.Now(),
	}

	if err := database.DB.Create(&timeEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create time entry"})
		return
	}

	// Reload the time entry with project data
	if err := database.DB.Preload("Project").First(&timeEntry, timeEntry.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch created time entry"})
		return
	}

	response := models.TimeEntryResponse{
		ID:        timeEntry.ID,
		StartTime: timeEntry.StartTime,
		EndTime:   timeEntry.EndTime,
		Duration:  timeEntry.Duration,
		CreatedAt: timeEntry.CreatedAt,
	}

	// Add project data if available
	if timeEntry.Project != nil {
		response.Project = &models.ProjectResponse{
			ID:          timeEntry.Project.ID,
			Name:        timeEntry.Project.Name,
			Description: timeEntry.Project.Description,
			Color:       timeEntry.Project.Color,
			CreatedAt:   timeEntry.Project.CreatedAt,
		}
	}

	c.JSON(http.StatusCreated, response)
}

func GetTimeEntries(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(uuid.UUID)

	// Parse pagination parameters
	page := 1
	limit := 10 // default limit

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	log.Printf("GetTimeEntries pagination: page=%d, limit=%d, offset=%d", page, limit, (page-1)*limit)

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	var total int64
	if err := database.DB.Model(&models.TimeEntry{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count time entries"})
		return
	}

	// Get paginated time entries
	var timeEntries []models.TimeEntry
	if err := database.DB.Preload("Project").Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&timeEntries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch time entries"})
		return
	}

	// Convert to response format
	var data []models.TimeEntryResponse
	for _, entry := range timeEntries {
		response := models.TimeEntryResponse{
			ID:        entry.ID,
			StartTime: entry.StartTime,
			EndTime:   entry.EndTime,
			Duration:  entry.Duration,
			CreatedAt: entry.CreatedAt,
		}

		// Add project data if available
		if entry.Project != nil {
			response.Project = &models.ProjectResponse{
				ID:          entry.Project.ID,
				Name:        entry.Project.Name,
				Description: entry.Project.Description,
				Color:       entry.Project.Color,
				CreatedAt:   entry.Project.CreatedAt,
			}
		}

		data = append(data, response)
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := models.PaginatedTimeEntriesResponse{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func GetTimeEntry(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var timeEntry models.TimeEntry
	if err := database.DB.Preload("Project").Where("id = ? AND user_id = ?", id, userID).First(&timeEntry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		return
	}

	response := models.TimeEntryResponse{
		ID:        timeEntry.ID,
		StartTime: timeEntry.StartTime,
		EndTime:   timeEntry.EndTime,
		Duration:  timeEntry.Duration,
		CreatedAt: timeEntry.CreatedAt,
	}

	// Add project data if available
	if timeEntry.Project != nil {
		response.Project = &models.ProjectResponse{
			ID:          timeEntry.Project.ID,
			Name:        timeEntry.Project.Name,
			Description: timeEntry.Project.Description,
			Color:       timeEntry.Project.Color,
			CreatedAt:   timeEntry.Project.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

func UpdateTimeEntry(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.TimeEntryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var timeEntry models.TimeEntry
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&timeEntry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		return
	}

	// Update fields
	if req.ProjectID != nil {
		timeEntry.ProjectID = req.ProjectID
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end time format"})
			return
		}
		// Validate that end time is after start time
		if endTime.Before(timeEntry.StartTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "End time cannot be before start time"})
			return
		}
		timeEntry.EndTime = &endTime
		timeEntry.Duration = int64(endTime.Sub(timeEntry.StartTime).Seconds())
	}

	if err := database.DB.Save(&timeEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update time entry"})
		return
	}

	// Reload the time entry with project data
	if err := database.DB.Preload("Project").First(&timeEntry, timeEntry.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated time entry"})
		return
	}

	response := models.TimeEntryResponse{
		ID:        timeEntry.ID,
		StartTime: timeEntry.StartTime,
		EndTime:   timeEntry.EndTime,
		Duration:  timeEntry.Duration,
		CreatedAt: timeEntry.CreatedAt,
	}

	// Add project data if available
	if timeEntry.Project != nil {
		response.Project = &models.ProjectResponse{
			ID:          timeEntry.Project.ID,
			Name:        timeEntry.Project.Name,
			Description: timeEntry.Project.Description,
			Color:       timeEntry.Project.Color,
			CreatedAt:   timeEntry.Project.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

func StopTimeEntry(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var timeEntry models.TimeEntry
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&timeEntry).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		return
	}

	if timeEntry.EndTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time entry already stopped"})
		return
	}

	now := time.Now()
	timeEntry.EndTime = &now
	timeEntry.Duration = int64(now.Sub(timeEntry.StartTime).Seconds())

	if err := database.DB.Save(&timeEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop time entry"})
		return
	}

	// Reload the time entry with project data
	if err := database.DB.Preload("Project").First(&timeEntry, timeEntry.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stopped time entry"})
		return
	}

	response := models.TimeEntryResponse{
		ID:        timeEntry.ID,
		StartTime: timeEntry.StartTime,
		EndTime:   timeEntry.EndTime,
		Duration:  timeEntry.Duration,
		CreatedAt: timeEntry.CreatedAt,
	}

	// Add project data if available
	if timeEntry.Project != nil {
		response.Project = &models.ProjectResponse{
			ID:          timeEntry.Project.ID,
			Name:        timeEntry.Project.Name,
			Description: timeEntry.Project.Description,
			Color:       timeEntry.Project.Color,
			CreatedAt:   timeEntry.Project.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

func DeleteTimeEntry(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.TimeEntry{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete time entry"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Time entry deleted successfully"})
}
