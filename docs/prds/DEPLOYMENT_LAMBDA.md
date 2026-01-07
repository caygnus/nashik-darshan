# AWS Lambda Deployment Guide

This guide explains how to deploy the Nashik Darshan API to AWS Lambda for free (within AWS Free Tier limits).

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Architecture Overview](#architecture-overview)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [Deployment Steps](#deployment-steps)
6. [Cost Optimization](#cost-optimization)
7. [Troubleshooting](#troubleshooting)
8. [Local Testing](#local-testing)

## Prerequisites

### Required Tools

1. **AWS Account** - [Sign up here](https://aws.amazon.com/)
2. **AWS CLI** - For credentials management

   ```bash
   # Install AWS CLI
   curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
   sudo installer -pkg AWSCLIV2.pkg -target /

   # Configure AWS credentials
   aws configure
   ```

3. **Serverless Framework** - For deployment

   ```bash
   npm install -g serverless
   ```

4. **Go 1.24+** - Already required for development

### AWS Permissions

Your AWS IAM user needs the following permissions:

- Lambda (create, update, invoke functions)
- API Gateway (create, update HTTP APIs)
- CloudFormation (stack operations)
- S3 (for deployment artifacts)
- IAM (create execution roles)
- CloudWatch Logs (for logging)

## Architecture Overview

```
┌──────────────┐
│  API Gateway │  ← HTTPS Requests
│  (HTTP API)  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  AWS Lambda  │  ← Go Binary (bin/bootstrap)
│  (ARM64)     │  ← Deployment Mode: lambda
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Gin Router  │  ← Same router as local mode
│  (All Routes)│
└──────────────┘
```

### Key Features

- **Same Codebase**: No code duplication between local and Lambda deployments
- **Mode-Based Configuration**: Set `CAYGNUS_DEPLOYMENT_MODE=lambda`
- **Dependency Injection**: Uses fx for consistent setup across modes
- **Cost Efficient**: ARM64 Lambda for 20% cost savings
- **Optimized Binary**: Stripped symbols, minimal size for faster cold starts

## Quick Start

```bash
# 1. Build the Lambda binary
make build-lambda

# 2. Deploy to AWS
make deploy-lambda

# Output will show your API endpoint:
# https://xxxxx.execute-api.us-east-1.amazonaws.com/
```

## Configuration

### Environment Variables

Configure these in `serverless.yml` under `provider.environment`:

```yaml
environment:
  CAYGNUS_DEPLOYMENT_MODE: lambda # Required: Enable Lambda mode
  CAYGNUS_SERVER_ENV: prod # Environment: dev/prod/local

  # Database Configuration
  CAYGNUS_POSTGRES_HOST: ${env:POSTGRES_HOST}
  CAYGNUS_POSTGRES_PORT: ${env:POSTGRES_PORT}
  CAYGNUS_POSTGRES_USER: ${env:POSTGRES_USER}
  CAYGNUS_POSTGRES_PASSWORD: ${env:POSTGRES_PASSWORD}
  CAYGNUS_POSTGRES_DBNAME: ${env:POSTGRES_DBNAME}
  CAYGNUS_POSTGRES_SSLMODE: require

  # Supabase Configuration
  CAYGNUS_SUPABASE_URL: ${env:SUPABASE_URL}
  CAYGNUS_SUPABASE_PUBLISHABLE_KEY: ${env:SUPABASE_PUBLISHABLE_KEY}
  CAYGNUS_SUPABASE_SECRET_KEY: ${env:SUPABASE_SECRET_KEY}

  # Secrets
  CAYGNUS_SECRETS_ENCRYPTION_KEY: ${env:ENCRYPTION_KEY}
```

### Database Connection

**Important**: Lambda functions are stateless and may have multiple concurrent instances. Consider:

1. **Connection Pooling**: Adjust pool sizes for Lambda

   ```yaml
   CAYGNUS_POSTGRES_MAX_OPEN_CONNS: 5 # Lower than local (default: 10)
   CAYGNUS_POSTGRES_MAX_IDLE_CONNS: 2 # Lower than local (default: 5)
   CAYGNUS_POSTGRES_CONN_MAX_LIFETIME_MINUTES: 5 # Shorter lifetime
   ```

2. **Use RDS Proxy** (Recommended): Manages connection pooling

   - Reduces Lambda cold start times
   - Handles connection limits efficiently
   - [AWS RDS Proxy Setup Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy.html)

3. **Use Serverless Databases**: Consider:
   - AWS Aurora Serverless
   - Neon (serverless Postgres)
   - Supabase (already configured)

## Deployment Steps

### Step 1: Set Up Environment Variables

Create a `.env.production` file (don't commit this!):

```bash
# Database
POSTGRES_HOST=your-db-host.com
POSTGRES_PORT=5432
POSTGRES_USER=your_user
POSTGRES_PASSWORD=your_password
POSTGRES_DBNAME=your_database

# Supabase
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_PUBLISHABLE_KEY=your_publishable_key
SUPABASE_SECRET_KEY=your_secret_key

# Secrets
ENCRYPTION_KEY=your_encryption_key
```

### Step 2: Update serverless.yml

Review and customize `serverless.yml`:

```yaml
service: nashik-darshan-api # Change to your service name

provider:
  name: aws
  runtime: provided.al2 # Custom runtime for Go
  region: us-east-1 # Change to your preferred region
  architecture: arm64 # ARM64 for 20% cost savings
  memorySize: 512 # Adjust based on your needs (128-10240 MB)
  timeout: 30 # Max execution time (1-900 seconds)
```

### Step 3: Build and Deploy

```bash
# Build optimized Lambda binary
make build-lambda

# Deploy to AWS
make deploy-lambda
```

The deployment process will:

1. Build the Go binary for Linux/ARM64
2. Strip symbols to reduce size
3. Package the binary as `bin/bootstrap`
4. Upload to AWS S3
5. Create/update Lambda function
6. Configure API Gateway
7. Output your API endpoint

### Step 4: Verify Deployment

```bash
# Test the health endpoint
curl https://your-api-id.execute-api.us-east-1.amazonaws.com/health

# Expected response:
# {"status":"ok","timestamp":"..."}
```

### Step 5: View Logs

```bash
# View Lambda logs
serverless logs -f api

# Or use AWS CloudWatch console
# https://console.aws.amazon.com/cloudwatch/
```

## Cost Optimization

### AWS Free Tier (First 12 Months)

- **Lambda Requests**: 1M free requests/month
- **Lambda Compute**: 400,000 GB-seconds/month
- **API Gateway**: 1M API calls/month
- **CloudWatch Logs**: 5GB ingestion, 5GB storage

### Cost After Free Tier

With 512MB memory and ARM64 architecture:

| Metric        | Usage          | Cost/Month |
| ------------- | -------------- | ---------- |
| 100K requests | ~5 seconds avg | ~$0.50     |
| 1M requests   | ~5 seconds avg | ~$5.00     |
| 10M requests  | ~5 seconds avg | ~$50.00    |

### Optimization Tips

1. **Use ARM64 Architecture**: 20% cheaper than x86_64

   ```yaml
   architecture: arm64
   ```

2. **Optimize Memory**: Test and use minimum required

   ```yaml
   memorySize: 512 # Start here, adjust as needed
   ```

3. **Reduce Cold Starts**:

   - Keep binary size small (< 50MB)
   - Use provisioned concurrency for critical APIs (costs extra)
   - Consider Lambda SnapStart (when available for Go)

4. **Connection Pooling**: Use RDS Proxy to reduce connection overhead

5. **Caching**: Implement API caching in API Gateway
   ```yaml
   functions:
     api:
       events:
         - httpApi:
             path: /{proxy+}
             method: ANY
             cors: true
             caching:
               enabled: true
               ttlInSeconds: 300 # 5 minutes
   ```

## Troubleshooting

### Issue: "Task timed out after 30 seconds"

**Solution**: Increase timeout in `serverless.yml`:

```yaml
provider:
  timeout: 60 # Increase to 60 seconds
```

### Issue: "Out of memory"

**Solution**: Increase memory allocation:

```yaml
provider:
  memorySize: 1024 # Increase to 1GB
```

### Issue: Database connection errors

**Solutions**:

1. Check security groups allow Lambda to reach database
2. Use RDS Proxy for connection pooling
3. Reduce connection pool size in environment variables
4. Ensure database is accessible from Lambda's VPC

### Issue: Cold start latency

**Solutions**:

1. Keep binary size small (use `-ldflags="-s -w"`)
2. Use provisioned concurrency (costs extra)
3. Consider keeping Lambda "warm" with scheduled pings
4. Optimize database connection initialization

### Issue: "Invalid ELF header" or "exec format error"

**Solution**: Ensure you're building for Linux/ARM64:

```bash
GOOS=linux GOARCH=arm64 go build -o bin/bootstrap cmd/server/main.go
```

### Debug Deployment Issues

```bash
# View detailed deployment logs
serverless deploy --verbose

# Check Lambda function logs
serverless logs -f api --tail

# Test Lambda directly (bypass API Gateway)
serverless invoke -f api --log
```

## Local Testing

### Test Lambda Mode Locally

1. **Using SAM CLI**:

   ```bash
   # Install SAM CLI
   brew install aws-sam-cli

   # Build
   make build-lambda

   # Test locally
   sam local start-api

   # API will be available at http://localhost:3000
   ```

2. **Using Environment Variable**:

   ```bash
   # Set Lambda mode locally
   export CAYGNUS_DEPLOYMENT_MODE=lambda

   # Run the server (will start Lambda handler)
   go run cmd/server/main.go
   ```

### Test with Docker

```bash
# Build Lambda-compatible binary
make build-lambda

# Run in Lambda Runtime Interface Emulator
docker run -p 9000:8080 \
  -v $(pwd):/var/task \
  -e CAYGNUS_DEPLOYMENT_MODE=lambda \
  -e CAYGNUS_POSTGRES_HOST=host.docker.internal \
  public.ecr.aws/lambda/provided:al2-arm64 \
  bin/bootstrap
```

## Additional Resources

- [AWS Lambda Go Documentation](https://docs.aws.amazon.com/lambda/latest/dg/lambda-golang.html)
- [Serverless Framework Docs](https://www.serverless.com/framework/docs)
- [AWS Lambda Pricing](https://aws.amazon.com/lambda/pricing/)
- [API Gateway HTTP API Pricing](https://aws.amazon.com/api-gateway/pricing/)
- [Gin Lambda Adapter](https://github.com/awslabs/aws-lambda-go-api-proxy)

## Support

If you encounter issues:

1. Check CloudWatch Logs for error details
2. Review the [Troubleshooting](#troubleshooting) section
3. Open an issue in the project repository
4. Contact the development team

---

**Next Steps**: After deploying, consider setting up:

- Custom domain with Route53
- SSL certificate with ACM
- CI/CD pipeline with GitHub Actions
- Monitoring and alerts with CloudWatch
- API documentation with Swagger UI
