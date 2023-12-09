package user

import (
	"net/http"
	"pottogether/internal/auth"
	"pottogether/internal/response"
	"pottogether/pkg/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

type signUpRequest struct {
	Avatar int    `json:"avatar" binding:"required"`
	Email  string `json:"email" binding:"required"`
	Passwd string `json:"passwd" binding:"required"`
}

type loginRequest struct {
	Email  string `json:"email" binding:"required"`
	Passwd string `json:"passwd" binding:"required"`
}

func Signup(c *gin.Context) {
	// Create response
	r := response.New()

	// Parse request body to JSON format
	var signUpRequest signUpRequest
	if err := c.ShouldBindJSON(&signUpRequest); err != nil {
		logger.Error("[USER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Check pass in fields (Email has @ symbol)
	if !strings.Contains(signUpRequest.Email, "@") {
		logger.Warn("[USER] Invalid email address")
		r.Message = "Invalid email address"
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Register the user
	id, err := signUp(signUpRequest)
	if err != nil {
		r.Message = err.Error()
		if r.Message == "email already exists" {
			c.JSON(http.StatusBadRequest, r)
			return
		}
		c.JSON(http.StatusInternalServerError, r)
		return
	}
	token, err := auth.GenerateToken(id, signUpRequest.Email)
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = response.SignUpLoginResponse{ID: id, Token: token}
	c.JSON(http.StatusCreated, r)
}

func Login(c *gin.Context) {
	// Create response
	r := response.New()

	// Parse request body to JSON format
	var loginRequest loginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		logger.Warn("[USER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Login the user
	id, err := login(loginRequest)
	if err != nil {
		r.Message = err.Error()
		if r.Message == "user not found" {
			c.JSON(http.StatusNotFound, r)
			return
		} else if r.Message == "incorrect password" {
			c.JSON(http.StatusUnauthorized, r)
			return
		}
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Generate token
	token, err := auth.GenerateToken(id, loginRequest.Email)
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = response.SignUpLoginResponse{ID: id, Token: token}
	c.JSON(http.StatusOK, r)
}

func GetProfile(c *gin.Context) {
	// Create response
	r := response.New()

	// Get user info
	userInfo, err := getProfile(c.MustGet("UserID").(int))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			r.Message = "user not found"
			c.JSON(http.StatusNotFound, r)
			return
		}
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = userInfo
	c.JSON(http.StatusOK, r)
}

func GetToday(c *gin.Context) {
	// Create response
	r := response.New()

	// Get today's record
	todayRecord, err := getToday(c.MustGet("UserID").(int))
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = todayRecord
	c.JSON(http.StatusOK, r)
}

func GetInvterval(c *gin.Context) {
	// Create response
	r := response.New()

	// Get interval record
	intervalRecord, err := getInterval(c.MustGet("UserID").(int))
	if err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = intervalRecord
	c.JSON(http.StatusOK, r)
}