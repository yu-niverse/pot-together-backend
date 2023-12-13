package record

import (
	"net/http"
	"pottogether/internal/response"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb/query"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateRecord(c *gin.Context) {
	// Create response
	r := response.New()

	// Parse request body to JSON format
	var recordRequest query.RecordRequest
	if err := c.ShouldBindJSON(&recordRequest); err != nil {
		logger.Error("[RECORD] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	id, err := query.CreateRecord(recordRequest)
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

	record, err := query.GetRecordOverview(c.MustGet("id").(int), "record.user_id")
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
	record, err := query.GetRecordOverview(id, "record.room_id")
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
	record, err := query.GetRecordDetail(id)
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
	var doneRequest query.DoneRequest
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
	err = query.UpdateRecord(id, doneRequest)
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	c.JSON(http.StatusOK, r)
}
