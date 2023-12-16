package record

import (
	"fmt"
	"net/http"
	"pottogether/pkg/errhandler"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb/query"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateRecord(c *gin.Context) {
	// Parse request body to JSON format
	var req query.RecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	logger.Info("Request content: " + fmt.Sprintf("%+v", req))

	id, err := query.CreateRecord(req)
	if err != nil {
		errhandler.Error(c, err, "Error creating record")
		return
	}
	// Response
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"id":        id,
	})
}

func GetUserRecord(c *gin.Context) {
	record, err := query.GetRecordOverview(c.MustGet("id").(int), "record.user_id")
	if err != nil {
		errhandler.Error(c, err, "Error getting user record")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"data":      record,
		"message":   "Success",
	})
}

func GetRoomRecords(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	record, err := query.GetRecordOverview(id, "record.room_id")
	if err != nil {
		errhandler.Error(c, err, "Error getting room records")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"data":      record,
		"message":   "Success",
	})
}

func GetRecordDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	record, err := query.GetRecordDetail(id)
	if err != nil {
		errhandler.Error(c, err, "Error getting record detail")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"data":      record,
		"message":   "Success",
	})
}

func UpdateRecord(c *gin.Context) {
	// Parse request body to JSON format
	var req query.DoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	logger.Info("Request content: " + fmt.Sprintf("%+v", req))
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	err = query.UpdateRecord(id, req)
	if err != nil {
		errhandler.Error(c, err, "Error updating record")
		return
	}
	// Response
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"id":        id,
		"message":   "Successfully updated record with id: " + strconv.Itoa(id),
	})
}
