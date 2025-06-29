package main

import (
	"log"
	"net/http"
	"ticketing-system/config"
	"ticketing-system/controller"
	"ticketing-system/middleware"
	"ticketing-system/repository"
	"ticketing-system/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Ticketing System API
// @version 1.0
// @description A comprehensive ticketing system with user management, event management, and ticket sales
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description JWT token format: Bearer {token}

func main() {
	// Load configuration
	config.LoadConfig()

	// Set Gin mode
	gin.SetMode(config.AppConfig.Server.GinMode)

	// Connect to database
	config.ConnectDatabase()

	// Run migrations
	config.AutoMigrate()

	// Initialize dependencies
	userRepo := repository.NewUserRepository(config.DB)
	eventRepo := repository.NewEventRepository(config.DB)
	ticketRepo := repository.NewTicketRepository(config.DB)

	userService := service.NewUserService(
		userRepo,
		config.AppConfig.JWT.Secret,
		config.AppConfig.GetJWTDuration(),
	)
	eventService := service.NewEventService(eventRepo)
	ticketService := service.NewTicketService(ticketRepo, eventRepo, userRepo, config.DB)

	userController := controller.NewUserController(userService)
	eventController := controller.NewEventController(eventService)
	ticketController := controller.NewTicketController(ticketService)
	reportController := controller.NewReportController(ticketService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(userService)

	// Initialize Gin router
	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "ticketing-system",
			"version": "1.0.0",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("")
		{
			// Authentication routes
			public.POST("/register", userController.Register)
			public.POST("/login", userController.Login)

			// Public event routes
			public.GET("/events", eventController.GetAllEvents)
			public.GET("/events/:id", eventController.GetEventByID)
			public.GET("/events/active", eventController.GetActiveEvents)
			public.GET("/events/upcoming", eventController.GetUpcomingEvents)
		}

		// Protected routes (authentication required)
		protected := api.Group("")
		protected.Use(authMiddleware.AuthRequired())
		{
			// User profile routes
			protected.GET("/profile", userController.GetProfile)
			protected.PUT("/profile", userController.UpdateProfile)

			// Ticket routes for authenticated users
			protected.POST("/tickets", ticketController.BuyTicket)
			protected.GET("/tickets/my", ticketController.GetUserTickets)
			protected.GET("/tickets/:id", ticketController.GetTicketByID)
			protected.PATCH("/tickets/:id/cancel", ticketController.CancelTicket)
		}

		// Admin routes (admin access required)
		admin := api.Group("")
		admin.Use(authMiddleware.AuthRequired())
		admin.Use(authMiddleware.AdminRequired())
		{
			// User management (admin only)
			admin.GET("/users", userController.GetAllUsers)
			admin.DELETE("/users/:id", userController.DeleteUser)

			// Event management (admin only)
			admin.POST("/events", eventController.CreateEvent)
			admin.PUT("/events/:id", eventController.UpdateEvent)
			admin.DELETE("/events/:id", eventController.DeleteEvent)

			// Ticket management (admin only)
			admin.GET("/tickets", ticketController.GetAllTickets)
			admin.PATCH("/tickets/:id", ticketController.UpdateTicketStatus)

			// Reports (admin only)
			admin.GET("/reports/summary", reportController.GetSummaryReport)
			admin.GET("/reports/event/:id", reportController.GetEventReport)
		}
	}

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := ":" + config.AppConfig.Server.Port
	log.Printf("🚀 Server starting on port %s", config.AppConfig.Server.Port)
	log.Printf("📚 API Documentation available at http://localhost%s/swagger/index.html", port)
	log.Printf("🔍 Health check available at http://localhost%s/health", port)

	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 