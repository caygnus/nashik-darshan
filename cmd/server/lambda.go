package main

import (
	"context"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/omkar273/nashikdarshan/internal/logger"

	"go.uber.org/fx"
)

var (
	// ginLambda holds the Lambda adapter for the Gin router (HTTP API v2)
	ginLambda *ginadapter.GinLambdaV2
	// initOnce ensures the Lambda adapter is initialized only once
	initOnce sync.Once
)

// startLambdaServer initializes and starts the AWS Lambda handler
// The router is already initialized by fx dependency injection
func startLambdaServer(
	lc fx.Lifecycle,
	r *gin.Engine,
	log *logger.Logger,
) {
	log.Info("Configuring AWS Lambda handler")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Initializing Lambda adapter (HTTP API v2)...")

			// Initialize the Lambda adapter once using sync.Once
			// This ensures thread-safe initialization during cold starts
			initOnce.Do(func() {
				ginLambda = ginadapter.NewV2(r)
				log.Info("Lambda adapter initialized successfully")
			})

			// Start Lambda handler in a goroutine
			// Lambda runtime will block here waiting for invocations
			go func() {
				log.Info("Starting Lambda handler (waiting for invocations)...")
				lambda.Start(handleLambdaRequest)
			}()

			log.Info("Lambda handler ready")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Lambda handler stopped")
			return nil
		},
	})
}

// handleLambdaRequest processes incoming HTTP API v2 requests
func handleLambdaRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}
