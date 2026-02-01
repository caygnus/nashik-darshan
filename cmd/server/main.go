package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/omkar273/nashikdarshan/docs/swagger"
	"go.uber.org/fx"
)

// @title           Nashik Darshan API
// @version         1.0
// @description     API for Nashik Darshan
// @termsOfService  http://nashikdarshan.com/terms/

// @contact.name   API Support
// @contact.email  support@nashikdarshan.com

// @host      5p9ubi66hh.execute-api.ap-south-1.amazonaws.com
// @BasePath  /api/v1
// @schemes   https

// @securityDefinitions.apikey Authorization
// @in header
// @name Authorization
// @description JWT Bearer token, e.g. `Bearer <token>`.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for server-to-server or script access. Use X-API-Key header (alternative to Bearer).

func init() {
	// Set global timezone to UTC for consistent time handling
	time.Local = time.UTC
}

func main() {
	app := fx.New(
		// Infrastructure layer: config, logging, database, auth
		provideInfrastructure(),

		// Data layer: repositories
		provideRepositories(),

		// Business layer: services
		provideServices(),

		// Presentation layer: API handlers and router
		provideAPI(),

		// Application lifecycle: start appropriate server based on config
		fx.Invoke(startServer),
	)

	// Check if we're running in Lambda (via environment variable)
	// Lambda runtime sets AWS_LAMBDA_FUNCTION_NAME
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" || os.Getenv("CAYGNUS_DEPLOYMENT_MODE") == "lambda" {
		// For Lambda: start the app, then start the Lambda handler
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := app.Start(ctx); err != nil {
			panic(err)
		}
		defer app.Stop(ctx)

		// Start Lambda handler (this blocks)
		lambda.Start(handleLambdaRequest)
	} else {
		// For local/API mode: use normal fx lifecycle
		app.Run()
	}
}
