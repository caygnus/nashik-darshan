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
	"github.com/omkar273/nashikdarshan/internal/api/dto"
	"github.com/omkar273/nashikdarshan/internal/config"
	"github.com/omkar273/nashikdarshan/internal/domain/auth"
	"github.com/omkar273/nashikdarshan/internal/domain/user"
	"github.com/omkar273/nashikdarshan/internal/logger"
	"github.com/omkar273/nashikdarshan/internal/postgres"
	"github.com/omkar273/nashikdarshan/internal/repository"
	"github.com/omkar273/nashikdarshan/internal/service"
	"github.com/omkar273/nashikdarshan/internal/types"
	"github.com/samber/lo"
)

// OnboardUser handles user onboarding via CLI
// It checks if user exists in local DB, if exists skips Supabase auth (dev tool), otherwise signs up in Supabase
// Creates user in local DB and returns access token
func OnboardUser() error {
	// Create a new flag set for this command
	// Parse from os.Args, skipping the main flags (-cmd, -list, etc.)
	fs := flag.NewFlagSet("onboard-user", flag.ExitOnError)
	
	// Parse command-line flags
	var (
		email    = fs.String("email", "", "User email address (required)")
		password = fs.String("password", "", "User password (required for new users, optional if user exists)")
		name     = fs.String("name", "", "User name (optional, defaults to email prefix if not provided)")
	)
	
	// Parse from os.Args, but skip the program name and main flags
	// Find where "onboard-user" appears and parse from there
	args := os.Args[1:] // Skip program name
	startIdx := 0
	for i, arg := range args {
		if arg == "onboard-user" || (i > 0 && args[i-1] == "-cmd" && arg == "onboard-user") {
			startIdx = i + 1
			break
		}
	}
	
	// Parse remaining arguments (after -cmd onboard-user)
	if err := fs.Parse(args[startIdx:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required flags
	if *email == "" {
		return fmt.Errorf("email is required. Use --email flag")
	}

	// Generate default name from email if not provided
	userName := lo.TernaryF(*name != "", func() string { return *name }, func() string {
		// Extract part before @ as default name and capitalize first letter
		parts := strings.Split(*email, "@")
		if len(parts) > 0 && len(parts[0]) > 0 {
			namePart := strings.ToLower(parts[0])
			if len(namePart) > 0 {
				return strings.ToUpper(namePart[:1]) + namePart[1:]
			}
		}
		return "User"
	})

	log.Printf("Onboarding user: %s", *email)

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger, err := logger.NewLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Initialize Supabase auth helper (dev-only)
	supabaseAuth := NewSupabaseAuthHelper(cfg, logger)
	if supabaseAuth == nil {
		return fmt.Errorf("failed to create Supabase auth helper")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize Ent client first
	entClient, err := postgres.NewEntClient(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create Ent client: %w", err)
	}
	defer entClient.Close()

	// Wrap Ent client with transaction management
	dbClient := postgres.NewClient(entClient, logger)

	// Initialize repositories
	repoParams := repository.RepositoryParams{
		Client: dbClient,
		Logger: logger,
		Config: cfg,
	}
	userRepo := repository.NewUserRepository(repoParams)
	categoryRepo := repository.NewCategoryRepository(repoParams)
	placeRepo := repository.NewPlaceRepository(repoParams)
	reviewRepo := repository.NewReviewRepository(repoParams)

	// Initialize services
	serviceParams := service.ServiceParams{
		Logger:      logger,
		Config:      cfg,
		DB:          dbClient,
		UserRepo:    userRepo,
		CategoryRepo: categoryRepo,
		PlaceRepo:   placeRepo,
		ReviewRepo:  reviewRepo,
	}

	userService := service.NewUserService(serviceParams)

	// Check if user exists in local database first (by email)
	// If exists, skip password requirement (dev tool convenience)
	var existingUser *user.User
	var claims *auth.Claims
	var accessToken string

	log.Println("Checking local database for existing user...")
	existingUser, err = userRepo.GetByEmail(ctx, *email)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("failed to check user in local database: %w", err)
	}

	if existingUser != nil {
		// User exists in local DB - skip password requirement (dev tool)
		log.Printf("✅ User already exists in local database: %s (ID: %s)", *email, existingUser.ID)
		
		// If password provided, try to get token from Supabase
		if *password != "" {
			log.Println("Password provided, attempting to get access token from Supabase...")
			claims, accessToken, err = supabaseAuth.SignInWithPassword(ctx, *email, *password)
			if err != nil {
				log.Printf("⚠️  Warning: Failed to get access token from Supabase: %v", err)
				log.Println("Continuing without access token (dev tool mode)")
				// Create dummy claims for output
				claims = &auth.Claims{
					UserID: existingUser.ID,
					Email:  existingUser.Email,
				}
				accessToken = "" // No token available
			} else {
				log.Printf("✅ Got access token from Supabase")
			}
		} else {
			log.Println("No password provided - skipping Supabase authentication (dev tool mode)")
			// Create claims from existing user
			claims = &auth.Claims{
				UserID: existingUser.ID,
				Email:  existingUser.Email,
			}
			accessToken = "" // No token available
		}

		// Update name if provided and different
		if userName != existingUser.Name && *name != "" {
			existingUser.Name = userName
			existingUser.UpdatedAt = time.Now().UTC()
			err = userRepo.Update(ctx, existingUser)
			if err != nil {
				log.Printf("⚠️  Warning: Failed to update user name: %v", err)
			} else {
				log.Printf("✅ Updated user name to: %s", userName)
			}
		}
	} else {
		// User doesn't exist - require password for Supabase signup
		if *password == "" {
			return fmt.Errorf("password is required for new users. Use --password flag")
		}

		log.Println("User not found in local database, creating in Supabase...")
		
		// Try to sign in first (user might exist in Supabase but not local DB)
		claims, accessToken, err = supabaseAuth.SignInWithPassword(ctx, *email, *password)
		if err != nil {
			// User doesn't exist in Supabase, sign them up
			log.Println("User not found in Supabase, signing up...")
			claims, accessToken, err = supabaseAuth.SignUpWithPassword(ctx, *email, *password, userName)
			if err != nil {
				return fmt.Errorf("failed to sign up user: %w", err)
			}
			log.Printf("✅ User created in Supabase: %s (ID: %s)", *email, claims.UserID)
		} else {
			log.Printf("✅ User authenticated in Supabase: %s (ID: %s)", *email, claims.UserID)
		}

		// Create user in local database
		log.Println("Creating user in local database...")
		now := time.Now().UTC()
		newUser := &user.User{
			ID:    claims.UserID,
			Email: *email,
			Name:  userName,
			Role:  types.UserRoleUser,
			BaseModel: types.BaseModel{
				Status:    types.StatusPublished,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		existingUser, err = userService.Create(ctx, newUser)
		if err != nil {
			return fmt.Errorf("failed to create user in local database: %w", err)
		}
		log.Printf("✅ User created in local database: %s", existingUser.ID)
	}

	// Run onboarding
	onboardingService := service.NewOnboardingService(serviceParams)
	onboardingReq := &dto.OnboardingRequest{
		User: *existingUser,
	}
	err = onboardingService.Onboard(ctx, onboardingReq)
	if err != nil {
		return fmt.Errorf("failed to onboard user: %w", err)
	}
	log.Println("✅ User onboarding completed")

	// Print results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ User onboarding successful!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("User ID: %s\n", claims.UserID)
	fmt.Printf("Email: %s\n", *email)
	fmt.Printf("Name: %s\n", userName)
	if accessToken != "" {
		fmt.Println("\nAccess Token:")
		fmt.Println(accessToken)
	} else {
		fmt.Println("\n⚠️  No access token available (user exists, password not provided)")
		fmt.Println("To get access token, run with --password flag")
	}
	fmt.Println(strings.Repeat("=", 60))

	return nil
}
