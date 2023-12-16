package user

import (
	"fmt"
	"net/http"
	"pottogether/internal/auth"
	"pottogether/pkg/errhandler"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb/query"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SignUpRequest struct {
	Avatar   int    `json:"avatar" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"passwd" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"passwd" binding:"required"`
}

func Signup(c *gin.Context) {
	// Parse request body to JSON format
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	logger.Info("Request content: " + fmt.Sprintf("%+v", req))
	// Check email format
	if !strings.Contains(req.Email, "@") {
		errhandler.Info(c, fmt.Errorf("invalid email format"), "Error checking email format")
		return
	}
	// Check if email already exists
	exists, err := query.CheckEmail(req.Email)
	if err != nil {
		errhandler.Error(c, err, "Error checking email existence")
		return
	} else if exists {
		errhandler.Info(c, fmt.Errorf("email %s already exists", req.Email), "Error checking email existence")
		return
	}
	// User struct
	user := query.User{
		ID:       -1,
		Avatar:   req.Avatar,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	// Register the user
	id, err := query.SignUp(user)
	if err != nil {
		errhandler.Error(c, err, "Error registering user")
		return
	}
	// Response
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"data":      id,
		"message":   "Successfully registered user with email: " + req.Email,
	})
}

func Login(c *gin.Context) {
	// Parse request body to JSON format
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errhandler.Info(c, err, "Invalid request format")
		return
	}
	logger.Info("Request content: " + fmt.Sprintf("%+v", req))
	// Login the user
	id, err := query.Login(req.Email, req.Password)
	if err != nil {
		errhandler.Unauthorized(c, err, "Error logging in user")
		return
	} else if id == -1 {
		errhandler.Unauthorized(c, fmt.Errorf("invalid email or password"), "Error logging in user")
		return
	}
	// Generate token
	token, err := auth.GenerateToken(id, req.Email)
	if err != nil {
		errhandler.Error(c, err, "Error generating token")
		return
	}
	// Response
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"data":      token,
		"message":   "Successfully logged in user with email: " + req.Email,
	})
}

func GetProfile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		errhandler.Info(c, err, "Invalid userID")
		return
	}
	// Check if user exists
	exists, err := query.CheckUser(id)
	if err != nil {
		errhandler.Error(c, err, "Error checking user existence")
		return
	} else if !exists {
		errhandler.Info(c, fmt.Errorf("user with id %d does not exist", id), "Error checking user existence")
		return
	}
	// Get user info
	userProfile, err := query.GetProfile(id)
	if err != nil {
		errhandler.Error(c, err, "Error getting user profile")
		return
	}
	// Response
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"data":      userProfile,
		"message":   "Successfully retrieved user profile",
	})
}

func GetOverview(c *gin.Context) {
	id := c.GetInt("id")
	if id == 0 {
		errhandler.Info(c, nil, "Invalid userID")
		return
	}
	// Get user overview
	userOverview, err := query.GetOverview(id)
	if err != nil {
		errhandler.Error(c, err, "Error getting user overview")
		return
	}
	// Response
	c.JSON(http.StatusOK, gin.H{
		"isSuccess": true,
		"data":      userOverview,
		"message":   "Successfully retrieved user overview",
	})
}
