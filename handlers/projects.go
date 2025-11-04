package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time-tracker/database"
	"time-tracker/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateProject(c *gin.Context) {
	var req models.ProjectCreateRequest
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

	// Validate project name - don't allow "General"
	if req.Name == "General" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name 'General' is reserved"})
		return
	}

	// Set default color if not provided
	color := req.Color
	if color == "" {
		color = "#3B82F6" // Default blue color
	}

	project := models.Project{
		UserID:      userID.(uuid.UUID),
		Name:        req.Name,
		Description: req.Description,
		Color:       color,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, models.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Color:       project.Color,
		CreatedAt:   project.CreatedAt,
	})
}

func GetProjects(c *gin.Context) {
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

	log.Printf("GetProjects pagination: page=%d, limit=%d, offset=%d", page, limit, (page-1)*limit)

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	var total int64
	if err := database.DB.Model(&models.Project{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count projects"})
		return
	}

	// Get paginated time entries
	var projects []models.Project
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	// Convert to response format
	var data []models.ProjectResponse
	for _, entry := range projects {
		data = append(data, models.ProjectResponse{
			ID:          entry.ID,
			Name:        entry.Name,
			Description: entry.Description,
			Color:       entry.Color,
			CreatedAt:   entry.CreatedAt,
		})
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := models.PaginatedProjectResponse{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func GetProject(c *gin.Context) {
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

	var project models.Project
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, models.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Color:       project.Color,
		CreatedAt:   project.CreatedAt,
	})
}

func UpdateProject(c *gin.Context) {
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

	var req models.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.Project
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Update fields
	if req.Name != "" {
		// Validate project name - don't allow "General"
		if req.Name == "General" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project name 'General' is reserved"})
			return
		}
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Color != "" {
		project.Color = req.Color
	}

	if err := database.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, models.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Color:       project.Color,
		CreatedAt:   project.CreatedAt,
	})
}

func DeleteProject(c *gin.Context) {
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

	// Start a database transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// First, verify the project exists and belongs to the user
	var project models.Project
	if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&project).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Set project_id to null for all time entries associated with this project
	if err := tx.Model(&models.TimeEntry{}).
		Where("project_id = ? AND user_id = ?", id, userID).
		Update("project_id", nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update time entries"})
		return
	}

	// Now delete the project
	if err := tx.Delete(&project).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully and time entries updated"})
}
