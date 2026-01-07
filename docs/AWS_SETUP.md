# AWS Setup Guide - Permissions & Costs

## Required IAM Permissions

Your AWS IAM user needs the following permissions to deploy the Lambda function. You can either:

1. **Attach the provided IAM policy** (recommended for development)
2. **Use AWS managed policies** (less secure, more permissions)
3. **Create a custom policy** with minimal permissions

### Option 1: Attach Custom Policy (Recommended)

1. Go to [AWS IAM Console](https://console.aws.amazon.com/iam/)
2. Click **Users** → Select your user (`omkar-local-deploy`)
3. Click **Add permissions** → **Create inline policy**
4. Click **JSON** tab
5. Copy the contents from `docs/aws-iam-policy.json`
6. Click **Review policy**
7. Name it: `NashikDarshanLambdaDeployment`
8. Click **Create policy**

### Option 2: Use AWS Managed Policies (Quick but Less Secure)

Attach these managed policies to your IAM user:

- `AWSLambda_FullAccess`
- `AmazonAPIGatewayAdministrator`
- `CloudFormationFullAccess`
- `AmazonS3FullAccess`
- `IAMFullAccess` (or create a custom policy for role creation only)
- `CloudWatchLogsFullAccess`

**Note**: These policies grant more permissions than needed, but they're quick to set up for development.

### Option 3: Minimal Permissions (Most Secure)

If you want to create a minimal policy, you need permissions for:

- **Lambda**: Create, update, delete functions
- **API Gateway**: Create and manage HTTP APIs
- **CloudFormation**: Create and manage stacks
- **S3**: Create bucket and upload deployment artifacts
- **IAM**: Create execution roles for Lambda
- **CloudWatch Logs**: Create log groups
- **SSM Parameter Store**: Store deployment bucket name (optional - can be bypassed)

## Cost Breakdown

### 🆓 AWS Free Tier (First 12 Months)

**Completely FREE** if you stay within these limits:

| Service             | Free Tier Limit                                          | What It Means                                       |
| ------------------- | -------------------------------------------------------- | --------------------------------------------------- |
| **Lambda**          | 1M requests/month<br>400,000 GB-seconds/month            | ~3.3M requests/day<br>~13,333 requests/day at 512MB |
| **API Gateway**     | 1M API calls/month                                       | ~33K calls/day                                      |
| **CloudWatch Logs** | 5GB ingestion/month<br>5GB storage                       | ~167MB/day ingestion                                |
| **S3**              | 5GB storage<br>20,000 GET requests<br>2,000 PUT requests | Deployment artifacts only                           |

**For a small API (< 1000 requests/day)**: **$0/month** ✅

### 💰 Cost After Free Tier

With your current configuration (512MB ARM64 Lambda):

| Usage Level                        | Monthly Cost Estimate |
| ---------------------------------- | --------------------- |
| **Low** (10K requests/month)       | ~$0.05                |
| **Medium** (100K requests/month)   | ~$0.50                |
| **High** (1M requests/month)       | ~$5.00                |
| **Very High** (10M requests/month) | ~$50.00               |

**Cost Components:**

- **Lambda**: $0.20 per 1M requests + $0.0000166667 per GB-second (ARM64 is 20% cheaper)
- **API Gateway**: $1.00 per 1M requests
- **CloudWatch Logs**: $0.50 per GB ingested
- **S3**: $0.023 per GB stored (negligible for deployment artifacts)

### 💡 Cost Optimization Tips

1. **Use ARM64** (already configured) - 20% cheaper than x86_64 ✅
2. **Optimize memory** - Start with 512MB, reduce if possible
3. **Monitor usage** - Set up CloudWatch billing alerts
4. **Use caching** - Reduce API calls with API Gateway caching
5. **Optimize binary size** - Already done with `-ldflags="-s -w"` ✅

## Setting Up Permissions

### Step 1: Create IAM Policy

1. Open AWS Console → IAM → Policies
2. Click **Create policy**
3. Go to **JSON** tab
4. Copy policy from `docs/aws-iam-policy.json`
5. Click **Next**
6. Name: `NashikDarshanLambdaDeployment`
7. Click **Create policy**

### Step 2: Attach Policy to User

1. Go to IAM → Users → `omkar-local-deploy`
2. Click **Add permissions** → **Attach policies directly**
3. Search for `NashikDarshanLambdaDeployment`
4. Select it and click **Next**
5. Click **Add permissions**

### Step 3: Verify Permissions

```bash
# Test if you can create an S3 bucket
aws s3 mb s3://nashik-darshan-api-deployments-ap-south-1 --region ap-south-1

# Test if you can call STS (should work)
aws sts get-caller-identity

# Clean up test bucket
aws s3 rb s3://nashik-darshan-api-deployments-test
```

## Alternative: Use Deployment Bucket Configuration

If you don't want to grant SSM permissions, you can specify a deployment bucket directly in `serverless.yml`:

```yaml
provider:
  deploymentBucket:
    name: nashik-darshan-api-deployments-ap-south-1
    blockPublicAccess: true
```

Then create the bucket manually:

```bash
aws s3 mb s3://nashik-darshan-api-deployments-ap-south-1 --region ap-south-1
```

This bypasses the SSM parameter requirement.

## Monitoring Costs

### Set Up Billing Alerts

1. Go to [AWS Billing Console](https://console.aws.amazon.com/billing/)
2. Click **Billing preferences**
3. Enable **Receive Billing Alerts**
4. Go to **CloudWatch** → **Alarms** → **Create alarm**
5. Select **Billing** metric
6. Set threshold (e.g., $5.00)
7. Add email notification

### View Current Usage

```bash
# Check Lambda invocations
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Invocations \
  --dimensions Name=FunctionName,Value=nashik-darshan-api-dev-api \
  --start-time $(date -u -d '1 day ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 3600 \
  --statistics Sum
```

## Summary

✅ **Permissions Needed**: See `docs/aws-iam-policy.json`  
✅ **Cost**: **FREE** for first 12 months (within limits)  
✅ **After Free Tier**: ~$0.50/month for 100K requests  
✅ **Setup Time**: ~5 minutes to attach IAM policy

**Next Steps:**

1. Attach the IAM policy to your user
2. Create the S3 deployment bucket (or let Serverless create it)
3. Run `make deploy-lambda`
