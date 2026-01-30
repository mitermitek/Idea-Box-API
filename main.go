package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Box struct {
	ID          uint
	Title       string
	Description string
	Ideas       []Idea
}

type Idea struct {
	ID          uint
	Title       string
	Description string
	BoxID       int
	Box         Box
}

type BoxRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type IdeaRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type BoxResponse struct {
	ID          uint           `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Ideas       []IdeaResponse `json:"ideas"`
}

type IdeaResponse struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("idea-box.db"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	db.AutoMigrate(&Box{}, &Idea{})

	router := gin.Default()

	router.GET("/boxes", getBoxes)
	router.GET("/boxes/:id", getBox)
	router.POST("/boxes", createBox)
	router.PUT("/boxes/:id", updateBox)
	router.DELETE("/boxes/:id", deleteBox)

	router.GET("/boxes/:id/ideas", getBoxIdeas)
	router.GET("/boxes/:id/ideas/:ideaId", getBoxIdea)
	router.POST("/boxes/:id/ideas", createBoxIdea)
	router.PUT("/boxes/:id/ideas/:ideaId", updateBoxIdea)
	router.DELETE("/boxes/:id/ideas/:ideaId", deleteBoxIdea)

	router.Run(":8080")
}

func getBoxes(c *gin.Context) {
	var boxes []Box

	result := db.Preload("Ideas").Find(&boxes)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	responses := make([]BoxResponse, 0)

	for _, box := range boxes {
		ideaResponses := make([]IdeaResponse, 0)
		for _, idea := range box.Ideas {
			ideaResponses = append(ideaResponses, IdeaResponse{
				ID:          idea.ID,
				Title:       idea.Title,
				Description: idea.Description,
			})
		}

		responses = append(responses, BoxResponse{
			ID:          box.ID,
			Title:       box.Title,
			Description: box.Description,
			Ideas:       ideaResponses,
		})
	}

	c.JSON(http.StatusOK, responses)
}

func getBox(c *gin.Context) {
	id := c.Param("id")
	var box Box

	result := db.Preload("Ideas").First(&box, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
		return
	}

	ideaResponses := make([]IdeaResponse, 0)
	for _, idea := range box.Ideas {
		ideaResponses = append(ideaResponses, IdeaResponse{
			ID:          idea.ID,
			Title:       idea.Title,
			Description: idea.Description,
		})
	}

	response := BoxResponse{
		ID:          box.ID,
		Title:       box.Title,
		Description: box.Description,
		Ideas:       ideaResponses,
	}

	c.JSON(http.StatusOK, response)
}

func createBox(c *gin.Context) {
	var request BoxRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	box := Box{Title: request.Title, Description: request.Description}
	result := db.Create(&box)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	response := BoxResponse{
		ID:          box.ID,
		Title:       box.Title,
		Description: box.Description,
		Ideas:       []IdeaResponse{},
	}

	c.JSON(http.StatusCreated, response)
}

func updateBox(c *gin.Context) {
	id := c.Param("id")

	var request BoxRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var box Box
	if err := db.First(&box, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
		return
	}

	result := db.Model(&box).Updates(BoxRequest{
		Title:       request.Title,
		Description: request.Description,
	})

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	response := BoxResponse{
		ID:          box.ID,
		Title:       box.Title,
		Description: box.Description,
		Ideas:       []IdeaResponse{},
	}

	c.JSON(http.StatusOK, response)
}

func deleteBox(c *gin.Context) {
	id := c.Param("id")

	var box Box
	if err := db.First(&box, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
		return
	}

	db.Where("box_id = ?", id).Delete(&Idea{})

	result := db.Delete(&box)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func getBoxIdeas(c *gin.Context) {
	boxID := c.Param("id")

	var box Box
	if err := db.First(&box, "id = ?", boxID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
		return
	}

	var ideas []Idea
	result := db.Where("box_id = ?", boxID).Find(&ideas)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	responses := make([]IdeaResponse, 0)
	for _, idea := range ideas {
		responses = append(responses, IdeaResponse{
			ID:          idea.ID,
			Title:       idea.Title,
			Description: idea.Description,
		})
	}

	c.JSON(http.StatusOK, responses)
}

func getBoxIdea(c *gin.Context) {
	boxID := c.Param("id")
	ideaID := c.Param("ideaId")

	var box Box
	if err := db.First(&box, "id = ?", boxID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
		return
	}

	var idea Idea
	if err := db.Where("id = ? AND box_id = ?", ideaID, boxID).First(&idea).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "idea not found"})
		return
	}

	response := IdeaResponse{
		ID:          idea.ID,
		Title:       idea.Title,
		Description: idea.Description,
	}

	c.JSON(http.StatusOK, response)
}

func createBoxIdea(c *gin.Context) {
	boxID := c.Param("id")

	var request IdeaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var box Box
	if err := db.First(&box, "id = ?", boxID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
		return
	}

	idea := Idea{
		Title:       request.Title,
		Description: request.Description,
		BoxID:       int(box.ID),
	}

	result := db.Create(&idea)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	response := IdeaResponse{
		ID:          idea.ID,
		Title:       idea.Title,
		Description: idea.Description,
	}

	c.JSON(http.StatusCreated, response)
}

func updateBoxIdea(c *gin.Context) {
	boxID := c.Param("id")
	ideaID := c.Param("ideaId")

	var request IdeaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var box Box
	if err := db.First(&box, "id = ?", boxID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
		return
	}

	var idea Idea
	if err := db.Where("id = ? AND box_id = ?", ideaID, boxID).First(&idea).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "idea not found"})
		return
	}

	result := db.Model(&idea).Updates(IdeaRequest{
		Title:       request.Title,
		Description: request.Description,
	})

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	response := IdeaResponse{
		ID:          idea.ID,
		Title:       idea.Title,
		Description: idea.Description,
	}

	c.JSON(http.StatusOK, response)
}

func deleteBoxIdea(c *gin.Context) {
	boxID := c.Param("id")
	ideaID := c.Param("ideaId")

	var box Box
	if err := db.First(&box, "id = ?", boxID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found"})
		return
	}

	var idea Idea
	if err := db.Where("id = ? AND box_id = ?", ideaID, boxID).First(&idea).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "idea not found"})
		return
	}

	result := db.Delete(&idea)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
