package router

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"

	"renfound_v1/config"
	"renfound_v1/infrastructure/auth"
	"renfound_v1/internal/delivery/http/handler"
	"renfound_v1/internal/delivery/http/middleware"
	"renfound_v1/internal/usecase/user"
	"renfound_v1/internal/utils/validator"
)

// Router handles routing for the application
type Router struct {
	app            *fiber.App
	cfg            *config.AppConfig
	userHandler    *handler.UserHandler
	authMiddleware *middleware.AuthMiddleware
	logMiddleware  *middleware.LoggingMiddleware
	logger         *zap.Logger
}

// NewRouter creates a new router
func NewRouter(
	cfg *config.AppConfig,
	userService user.Service,
	telegramAuth *auth.TelegramAuth,
) *Router {
	logger := cfg.Logger.With(zap.String("component", "router"))

	// Create validator
	validatorUtil := validator.NewValidator(logger)

	// Create handlers
	userHandler := handler.NewUserHandler(userService, validatorUtil, logger)

	// Create middlewares
	authMiddleware := middleware.NewAuthMiddleware(telegramAuth, logger)
	logMiddleware := middleware.NewLoggingMiddleware(logger)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		// Override default error handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Check if it's a Fiber error
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Return JSON error
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Register global middlewares
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))
	app.Use(logMiddleware.Logger())
	app.Use(logMiddleware.RecoverWithLogger())

	return &Router{
		app:            app,
		cfg:            cfg,
		userHandler:    userHandler,
		authMiddleware: authMiddleware,
		logMiddleware:  logMiddleware,
		logger:         logger,
	}
}

// SetupRoutes sets up the routes
func (r *Router) SetupRoutes() {
	api := r.app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/telegram", r.userHandler.AuthWithTelegram)
	auth.Post("/refresh", r.userHandler.RefreshTokens)
	auth.Post("/logout", r.userHandler.Logout)
	auth.Post("/logout-all", r.authMiddleware.Authenticate(), r.userHandler.LogoutAll)

	// User routes
	users := api.Group("/users", r.authMiddleware.Authenticate())
	users.Get("/me", r.userHandler.GetMe)
	users.Delete("/me", r.userHandler.DeleteMe)
}

// Start starts the server
func (r *Router) Start() error {
	r.logger.Info("Starting server", zap.String("host", r.cfg.Config.Server.Host), zap.String("port", r.cfg.Config.Server.Port))

	return r.app.Listen(r.cfg.Config.Server.Host + ":" + r.cfg.Config.Server.Port)
}

// Shutdown gracefully shuts down the server
func (r *Router) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down server with context")

	return r.app.ShutdownWithContext(ctx)
}
