package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"pottogether/api/record"
	"pottogether/api/user"
	"pottogether/config"
	"pottogether/internal/auth"
	"pottogether/pkg/logger"
	"pottogether/pkg/mariadb"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var err error

func API_init(LOG_PATH string) {
	// Load configuration
	if config.LoadConfig() == nil {
		fmt.Println("Error loading config file")
		return
	}
	// Init Logger
	logger.InitLogger(config.Viper.GetString(LOG_PATH))
	logger.Log.Info("Logger enabled, log file: " + config.Viper.GetString(LOG_PATH))
	// Init JWT
	auth.SetJWTKey()
	// Connect to MySQL
	if err = mariadb.Connect_init(); err != nil {
		logger.Error("Error connecting to mariadb: " + err.Error())
		return
	}
	logger.Info("MariaDB connected")
}

func Main() {
	API_init("API_LOG_FILE")
	ctx, cancel := context.WithCancel(context.Background())
	Quit := make(chan os.Signal, 1)

	// Gin Settings
	gin.SetMode(gin.ReleaseMode)
	f, _ := os.Create(config.Viper.GetString("API_GIN_LOG"))
	gin.DefaultWriter = io.MultiWriter(f)
	router := gin.Default()

	// CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Content-Type", "Accept", "Content-Length", "Authorization", "Origin", "X-Requested-With"}
	router.RedirectFixedPath = true
	router.Use(cors.New(corsConfig))
	router.Use(logger.GinLog())

	// Healthcheck
	router.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Healthcheck OK!")
	})

	// Signup and Login
	router.POST("users/signup", user.Signup)
	router.POST("users/login", user.Login)

	// Auth middleware for all routes below
	router.Use(auth.ValidateToken)

	// User Routes
	userGroup := router.Group("/users")
	userGroup.GET("/profile", user.GetProfile)
	userGroup.GET("/today", user.GetToday)
	userGroup.GET("/interval", user.GetInvterval)

	// Room Routes
	RoomGroup := router.Group("/rooms")
	RoomGroup.GET(":id/records", record.GetRoomRecords)

	// Ingredient Routes

	// Record Routes
	recordGroup := router.Group("/records")
	recordGroup.POST("", record.CreateRecord)
	recordGroup.GET("", record.GetUserRecord)
	recordGroup.GET("/:id", record.GetRecordDetail)
	recordGroup.PATCH("/:id", record.UpdateRecord)

	// Pot Routes

	// Start API service
	srv := &http.Server{
		Addr:    ":" + os.Args[1],
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error starting API server: " + err.Error())
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	signal.Notify(Quit, syscall.SIGINT, syscall.SIGTERM)
	<-Quit
	logger.Info("Shutting down API server...")
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down API server: " + err.Error())
		os.Exit(1)
	}
	mariadb.DB.Close()
	logger.Info("API server exited")
}
