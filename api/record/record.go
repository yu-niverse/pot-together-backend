package record

import (
	"fmt"
	"mime/multipart"
	"pottogether/internal/s3"
	"pottogether/pkg/errhandler"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb/query"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateRecordRequest struct {
	RoomID       int    `json:"roomID" binding:"required"`
	PotID        string `json:"potID" binding:"required"`
	IngredientID int    `json:"ingredientID" binding:"required"`
}

type UpdateRecordRequest struct {
	Image     *multipart.FileHeader `form:"image" binding:"required"`
	Caption   string                `form:"caption" binding:"required"`
	Interval  int                   `form:"interval" binding:"required"`
	Interrupt int                   `form:"interrupt"`
	Status    int                   `form:"status" binding:"required"`
}

func CreateRecord(c *gin.Context) {
	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	logger.Info("Request content: " + fmt.Sprintf("%+v", req))
	record := query.Record{
		ID:           -1,
		UserID:       c.GetInt("id"),
		RoomID:       req.RoomID,
		PotID:        req.PotID,
		IngredientID: req.IngredientID,
		Image:        c.GetString("image"),
		Caption:      "",
		Interval:     0,
		FinishTime:   0,
		Interrupt:    0,
		Status:       0,
	}
	recordID, err := query.CreateRecord(record)
	if err != nil {
		errhandler.Error(c, err, "Error creating record")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data": gin.H{
			"recordID": recordID,
		},
		"message": "Record created successfully",
	})
}

func UpdateRecord(c *gin.Context) {
	recordID, err := strconv.Atoi(c.Param("recordID"))
	s3.UploadMiddleware(c, "record", c.Param("recordID"))
	var req UpdateRecordRequest
	if err := c.ShouldBind(&req); err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	logger.Info("Request content: " + fmt.Sprintf("%+v", req))
	if err != nil {
		errhandler.Info(c, err, "Invalid recordID")
		return
	}
	record := query.Record{
		ID:           recordID,
		UserID:       -1,
		RoomID:       -1,
		PotID:        "",
		IngredientID: -1,
		Image:        c.GetString("image"),
		Caption:      req.Caption,
		Interval:     req.Interval,
		FinishTime:   int(time.Now().Unix()),
		Interrupt:    req.Interrupt,
		Status:       req.Status,
	}
	err = query.UpdateRecord(record)
	if err != nil {
		errhandler.Error(c, err, "Error updating record")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data":      nil,
		"message":   "Record updated successfully",
	})
}

func GetUserRecords(c *gin.Context) {
	records, err := query.GetUserRecords(c.GetInt("id"))
	if err != nil {
		errhandler.Error(c, err, "Error getting records")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data":      records,
		"message":   "Records retrieved successfully",
	})
}

func GetRecordDetail(c *gin.Context) {
	recordID, err := strconv.Atoi(c.Param("recordID"))
	if err != nil {
		errhandler.Info(c, err, "Invalid recordID")
		return
	}
	record, err := query.GetRecordDetail(recordID)
	if err != nil {
		errhandler.Error(c, err, "Error getting record detail")
		return
	}
	c.JSON(200, gin.H{
		"isSuccess": true,
		"data":      record,
		"message":   "Record detail retrieved successfully",
	})
}
