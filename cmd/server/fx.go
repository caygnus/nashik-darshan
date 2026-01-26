package main

import (
	"github.com/gin-gonic/gin"
	"github.com/omkar273/nashikdarshan/internal/api"
	v1 "github.com/omkar273/nashikdarshan/internal/api/v1"
	"github.com/omkar273/nashikdarshan/internal/auth"
	"github.com/omkar273/nashikdarshan/internal/config"
	"github.com/omkar273/nashikdarshan/internal/logger"
	"github.com/omkar273/nashikdarshan/internal/postgres"
	"github.com/omkar273/nashikdarshan/internal/repository"
	"github.com/omkar273/nashikdarshan/internal/security"
	"github.com/omkar273/nashikdarshan/internal/service"
	"github.com/omkar273/nashikdarshan/internal/validator"

	"go.uber.org/fx"
)

// provideInfrastructure provides all infrastructure dependencies
func provideInfrastructure() fx.Option {
	return fx.Provide(
		config.NewConfig,
		validator.NewValidator,
		logger.NewLogger,
		postgres.NewEntClient,
		postgres.NewClient,
		auth.NewSupabaseProvider,
	)
}

// provideRepositories provides all repository implementations
func provideRepositories() fx.Option {
	return fx.Provide(
		repository.NewUserRepository,
		repository.NewCategoryRepository,
		repository.NewPlaceRepository,
		repository.NewReviewRepository,
		repository.NewSecretRepository,
	)
}

// provideServices provides all business logic services
func provideServices() fx.Option {
	return fx.Provide(
		security.NewEncryptionService,
		service.NewAuthService,
		service.NewUserService,
		service.NewOnboardingService,
		service.NewCategoryService,
		service.NewPlaceService,
		service.NewReviewService,
		service.NewSecretService,
	)
}

// provideAPI provides API handlers and router
func provideAPI() fx.Option {
	return fx.Provide(
		provideHandlers,
		provideRouter,
	)
}

// provideHandlers creates API handlers from services
func provideHandlers(
	logger *logger.Logger,
	authService service.AuthService,
	userService service.UserService,
	categoryService service.CategoryService,
	placeService service.PlaceService,
	reviewService service.ReviewService,
	secretService service.SecretService,
) *api.Handlers {
	return &api.Handlers{
		Health:   v1.NewHealthHandler(logger),
		Auth:     v1.NewAuthHandler(authService),
		User:     v1.NewUserHandler(userService),
		Category: v1.NewCategoryHandler(categoryService),
		Place:    v1.NewPlaceHandler(placeService),
		Review:   v1.NewReviewHandler(reviewService),
		Secret:   v1.NewSecretHandler(secretService),
	}
}

// provideRouter creates the Gin router with handlers
func provideRouter(handlers *api.Handlers, cfg *config.Configuration, logger *logger.Logger, secretService service.SecretService) *gin.Engine {
	return api.NewRouter(handlers, cfg, logger, secretService)
}
