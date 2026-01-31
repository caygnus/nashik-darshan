package internal

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/omkar273/nashikdarshan/ent"
	"github.com/omkar273/nashikdarshan/internal/config"
	"github.com/omkar273/nashikdarshan/internal/logger"
	"github.com/omkar273/nashikdarshan/internal/postgres"
	"github.com/omkar273/nashikdarshan/internal/repository"
	"github.com/omkar273/nashikdarshan/internal/security"
	"github.com/omkar273/nashikdarshan/internal/service"
	"github.com/omkar273/nashikdarshan/internal/types"
)

// CreateAPIKey creates an API key for a user and prints the raw key.
// The key works with internal/rest/middleware (X-API-Key header).
// If neither --user-id nor --email is passed, types.DefaultUserID is used.
// Usage: go run scripts/main.go -cmd create-api-key [--name "etst key"] [--user-id <id> | --email <email>]
func CreateAPIKey() error {
	fs := flag.NewFlagSet("create-api-key", flag.ExitOnError)
	var (
		name   = fs.String("name", "etst key", "API key name (display label)")
		userID = fs.String("user-id", "", "User ID that will own the API key (default: types.DefaultUserID)")
		email  = fs.String("email", "", "User email to look up (alternative to --user-id)")
	)
	args := os.Args[1:]
	startIdx := 0
	for i, arg := range args {
		if arg == "create-api-key" || (i > 0 && args[i-1] == "-cmd" && arg == "create-api-key") {
			startIdx = i + 1
			break
		}
	}
	if err := fs.Parse(args[startIdx:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *userID != "" && *email != "" {
		return fmt.Errorf("provide only one of --user-id or --email")
	}

	log.Printf("Creating API key with name: %s", *name)

	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	lgr, err := logger.NewLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	encryptionSvc, err := security.NewEncryptionService(cfg, lgr)
	if err != nil {
		return fmt.Errorf("failed to create encryption service (check encryption_key in config): %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	entClient, err := postgres.NewEntClient(cfg, lgr)
	if err != nil {
		return fmt.Errorf("failed to create Ent client: %w", err)
	}
	defer entClient.Close()

	dbClient := postgres.NewClient(entClient, lgr)
	repoParams := repository.RepositoryParams{
		Client: dbClient,
		Logger: lgr,
		Config: cfg,
	}
	userRepo := repository.NewUserRepository(repoParams)
	secretRepo := repository.NewSecretRepository(repoParams)
	categoryRepo := repository.NewCategoryRepository(repoParams)
	placeRepo := repository.NewPlaceRepository(repoParams)
	reviewRepo := repository.NewReviewRepository(repoParams)

	resolvedUserID := types.DefaultUserID
	if *userID != "" {
		resolvedUserID = *userID
	} else if *email != "" {
		u, err := userRepo.GetByEmail(ctx, *email)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("no user found with email %q: %w", *email, err)
			}
			return fmt.Errorf("failed to look up user by email: %w", err)
		}
		resolvedUserID = u.ID
		log.Printf("Resolved user: %s (%s)", u.Email, resolvedUserID)
	} else {
		log.Printf("No --user-id or --email provided, using default user ID: %s", types.DefaultUserID)
	}

	serviceParams := service.ServiceParams{
		Logger:             lgr,
		Config:             cfg,
		DB:                 dbClient,
		UserRepo:           userRepo,
		CategoryRepo:       categoryRepo,
		PlaceRepo:          placeRepo,
		ReviewRepo:         reviewRepo,
		SecretRepo:         secretRepo,
		EncryptionService:  encryptionSvc,
	}
	secretService := service.NewSecretService(serviceParams)

	secretEntity, rawKey, err := secretService.CreateAPIKey(ctx, resolvedUserID, *name, types.SecretTypePrivateKey, nil)
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("API key created successfully")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Name:    %s\n", secretEntity.Name)
	fmt.Printf("ID:      %s\n", secretEntity.ID)
	fmt.Printf("Prefix:  %s\n", secretEntity.Prefix)
	fmt.Println()
	fmt.Println("API Key (use in X-API-Key header):")
	fmt.Println(rawKey)
	fmt.Println(strings.Repeat("=", 60))

	return nil
}
