package record

import (
	"net/http"
	"pottogether/internal/response"
	"pottogether/pkg/logger"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RecordRequest struct {
	UserID       int  `json:"user_id" binding:"required"`
	RoomID       int  `json:"room_id" binding:"required"`
	PotID        int  `json:"pot_id" binding:"required"`
	TimeInterval *int `json:"time_interval" binding:"required"`
	IngredientID int  `json:"ingredient_id" binding:"required"`
	Interrupt    *int `json:"interrupt" binding:"required"`
	Status       *int `json:"status" binding:"required"`
}

type DoneRequest struct {
	Image        string `json:"image" binding:"required"`
	Caption      string `json:"caption" binding:"required"`
	TimeInterval *int   `json:"time_interval" binding:"required"`
	Interrupt    *int   `json:"interrupt" binding:"required"`
	Status       *int   `json:"status" binding:"required"`
}

func CreateRecord(c *gin.Context) {
	// Create response
	r := response.New()

	// Parse request body to JSON format
	var recordRequest RecordRequest
	if err := c.ShouldBindJSON(&recordRequest); err != nil {
		logger.Error("[RECORD] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	id, err := createRecord(recordRequest)
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = response.RecordReponse{ID: id}
	c.JSON(http.StatusOK, r)
}

func GetUserRecord(c *gin.Context) {
	// Create response
	r := response.New()

	record, err := getRecordOverview(c.MustGet("id").(int), "record.user_id")
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = record
	c.JSON(http.StatusOK, r)
}

func GetRoomRecords(c *gin.Context) {
	// Create response
	r := response.New()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}
	record, err := getRecordOverview(id, "record.room_id")
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = record
	c.JSON(http.StatusOK, r)
}

func GetRecordDetail(c *gin.Context) {
	// Create response
	r := response.New()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}
	record, err := getRecordDetail(id)
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = record
	c.JSON(http.StatusOK, r)
}

func UpdateRecord(c *gin.Context) {
	// Create response
	r := response.New()

	// Parse request body to JSON format
	var doneRequest DoneRequest
	if err := c.ShouldBindJSON(&doneRequest); err != nil {
		logger.Error("[RECORD] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}
	err = updateRecord(id, doneRequest)
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	c.JSON(http.StatusOK, r)
}
