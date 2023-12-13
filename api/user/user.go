package user

import (
	"net/http"
	"pottogether/internal/auth"
	"pottogether/internal/response"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb/query"
	"strings"

	"github.com/gin-gonic/gin"
)

func Signup(c *gin.Context) {
	// Create response
	r := response.New()

	// Parse request body to JSON format
	var signUpRequest query.SignUpRequest
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
	id, err := query.SignUp(signUpRequest)
	if err != nil {
		r.Message = err.Error()
		if r.Message == "email already exists" {
			c.JSON(http.StatusBadRequest, r)
			return
		}
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	r.IsSuccess = true
	r.Data = response.SignUpResponse{ID: id}
	c.JSON(http.StatusCreated, r)
}

func Login(c *gin.Context) {
	// Create response
	r := response.New()

	// Parse request body to JSON format
	var loginRequest query.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		logger.Warn("[USER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Login the user
	id, err := query.Login(loginRequest)
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
	r.Data = response.LoginResponse{ID: id, Token: token}
	c.JSON(http.StatusOK, r)
}

func GetProfile(c *gin.Context) {
	// Create response
	r := response.New()

	// Get user info
	userInfo, err := query.GetProfile(c.MustGet("id").(int))
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
