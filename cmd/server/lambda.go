package main

import (
	"context"
	"sync"

	"github.com/aws/aws-lambda-go/events"
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
	// initMutex protects the initialization
	initMutex sync.Mutex
)

// startLambdaServer initializes and starts the AWS Lambda handler
// The router is already initialized by fx dependency injection
func startLambdaServer(
	lc fx.Lifecycle,
	r *gin.Engine,
	log *logger.Logger,
) {
	log.Info("Configuring AWS Lambda handler")

	// Initialize the Lambda adapter immediately (not in OnStart)
	// This ensures it's ready before Lambda runtime calls the handler
	initOnce.Do(func() {
		ginLambda = ginadapter.NewV2(r)
		log.Info("Lambda adapter initialized successfully")
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Lambda handler ready (HTTP API v2)")
			// Don't start lambda.Start here - it will be called by the Lambda runtime
			// The handler function will be invoked directly by Lambda
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
	// Ensure adapter is initialized (should already be, but safety check)
	initMutex.Lock()
	if ginLambda == nil {
		initMutex.Unlock()
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       `{"error":"Lambda adapter not initialized"}`,
		}, nil
	}
	adapter := ginLambda
	initMutex.Unlock()

	return adapter.ProxyWithContext(ctx, req)
}
