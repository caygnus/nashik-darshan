package main

import (
	"time"

	_ "github.com/omkar273/nashikdarshan/docs/swagger"
	"go.uber.org/fx"
)

// @title           Nashik Darshan API
// @version         1.0
// @description     API for Nashik Darshan
// @termsOfService  http://nashikdarshan.com/terms/

// @contact.name   API Support
// @contact.email  support@nashikdarshan.com

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Authorization
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. `Bearer <token>`.
// @type apiKey
// @required

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

	app.Run()
}
