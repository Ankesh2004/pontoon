# Local Testing Guide

This guide explains how to test Pontoon locally.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.23+ installed
- Git installed (for cloning test repos)

## Quick Start

### 1. Start Test Environment

```bash
./scripts/start-test-env.sh
```

This will:
- Start PostgreSQL and Redis in Docker containers
- Run database migrations
- Start the API server on port 8080
- Start the Worker process

### 2. Run API Tests

```bash
./scripts/test-api.sh
```

This will test:
- Health check endpoint
- Authentication flow
- Protected endpoints without tokens
- Webhook signature verification

### 3. Stop Test Environment

```bash
./scripts/stop-test-env.sh
```

## Manual Testing

### Health Check
```bash
curl http://localhost:8080/health
```

### GitHub OAuth Flow
1. Open browser: http://localhost:8080/auth/github
2. You'll be redirected to GitHub (will fail with test credentials, but shows the flow works)

### Test with Authentication
After getting a JWT token (from OAuth callback or manually):

```bash
# List projects
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/projects

# Create project
curl -X POST \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-app",
    "repo_url": "https://github.com/username/repo",
    "branch": "main"
  }' \
  http://localhost:8080/api/v1/projects

# Trigger deployment
curl -X POST \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"commit_sha": "abc123"}' \
  http://localhost:8080/api/v1/projects/PROJECT_ID/deploy

# Get deployment status
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/deployments/DEPLOYMENT_ID

# Get runtime logs
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/v1/deployments/DEPLOYMENT_ID/logs?tail=100"
```

### Environment Variables

```bash
# List env vars
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/projects/PROJECT_ID/env

# Create env var
curl -X POST \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key": "DATABASE_URL", "value": "postgres://..."}' \
  http://localhost:8080/api/v1/projects/PROJECT_ID/env

# Delete env var
curl -X DELETE \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/projects/PROJECT_ID/env/ENV_VAR_ID
```

### Webhook Testing

To test webhooks, you'll need to:
1. Create a project with a webhook secret
2. Use a tool like ngrok to expose your local API
3. Configure GitHub webhook to point to your ngrok URL

Example webhook payload:
```bash
curl -X POST \
  -H "X-Hub-Signature-256: sha256=CALCULATED_SIGNATURE" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "push",
    "ref": "refs/heads/main",
    "repository": {
      "full_name": "username/repo",
      "clone_url": "https://github.com/username/repo.git"
    },
    "head_commit": {
      "id": "abc123def456"
    }
  }' \
  "http://localhost:8080/webhooks/github?project_id=PROJECT_ID"
```

## Database Access

```bash
# Connect to PostgreSQL
docker exec -it pontoon-postgres psql -U pontoon -d pontoon

# Useful queries
SELECT * FROM users;
SELECT * FROM projects;
SELECT * FROM deployments;
SELECT * FROM env_vars;
```

## Redis Access

```bash
# Connect to Redis
docker exec -it pontoon-redis redis-cli

# Useful commands
KEYS *
GET deployment:DEPLOYMENT_ID:logs
```

## Troubleshooting

### Port Already in Use
If port 8080, 5432, or 6379 is already in use:
```bash
# Stop existing containers
./scripts/stop-test-env.sh

# Or manually
docker-compose -f docker-compose.test.yml down
```

### Migration Errors
```bash
# Reset database
docker-compose -f docker-compose.test.yml down -v
./scripts/start-test-env.sh
```

### API Not Starting
Check logs:
```bash
# View API logs
go run cmd/api/main.go

# Check if PostgreSQL is ready
docker-compose -f docker-compose.test.yml ps
```

## Next Steps

After testing, you can:
1. Create a GitHub OAuth App with real credentials
2. Update `.env` with real GitHub credentials
3. Test the full OAuth flow
4. Deploy a real application
