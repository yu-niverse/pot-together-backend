package ingredient

import (
	"mime/multipart"
	"pottogether/internal/s3"
	"pottogether/pkg/errhandler"
	"pottogether/pkg/mariadb/query"

	"github.com/gin-gonic/gin"
)

type AddIngredientRequest struct {
	Name        string                `form:"name" binding:"required"`
	Image       *multipart.FileHeader `form:"image" binding:"required"`
	Interval    int                   `form:"interval" binding:"required"`
	Requirement string                `form:"requirement"`
}

func GetIngredients(c *gin.Context) {
	ingredients, err := query.GetIngredients()
	if err != nil {
		errhandler.Error(c, err, "Error getting ingredients")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data":      ingredients,
		"message":   "Ingredients retrieved successfully",
	})
}

func AddIngredient(c *gin.Context) {
	s3.UploadMiddleware(c, "ingredient", "")
	var req AddIngredientRequest
	if err := c.ShouldBind(&req); err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	ingredient := query.Ingredient{
		ID:          -1,
		Name:        req.Name,
		Image:       c.GetString("image"),
		Interval:    req.Interval,
		Requirement: req.Requirement,
	}
	ingredientID, err := query.AddIngredient(ingredient)
	if err != nil {
		errhandler.Error(c, err, "Error adding ingredient")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data": gin.H{
			"ingredientID": ingredientID,
		},
		"message": "Ingredient added successfully",
	})
}
