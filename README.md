# Pontoon

A lightweight, event-driven Platform-as-a-Service (PaaS) engine built in Go. Pontoon automates the deployment, routing, and lifecycle management of containerized applications.

## Features

- **GitHub OAuth Authentication** - Secure login with GitHub
- **Multi-tenant Architecture** - User isolation with dedicated Docker networks
- **Async Deployment Pipeline** - Non-blocking builds with Redis/Asynq
- **Dynamic Routing** - Automatic Traefik integration for subdomain routing
- **Resource Management** - Memory/CPU limits per container
- **Webhook Support** - Auto-deploy on git push with HMAC verification
- **Real-time Logs** - WebSocket streaming of build logs

## Architecture

```
┌─────────────────────────────────────────────────┐
│         Cloudflare (DNS + Tunnel + WAF)         │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│  Traefik (dynamic routing, TLS, rate limiting)  │
└────┬──────────────┬──────────────┬──────────────┘
     │              │              │
┌────▼────┐  ┌─────▼─────┐  ┌────▼────────────┐
│ Go API  │  │ Go Worker │  │ User Containers │
│ (Chi)   │  │ (Asynq)   │  │ (isolated,      │
│         │  │           │  │  resource-limited)│
└────┬────┘  └─────┬─────┘  └─────────────────┘
     │              │
┌────▼──────────────▼─────┐
│ Postgres (Neon)         │
│ Redis (Upstash)         │
└─────────────────────────┘
```

## Prerequisites

- Go 1.23+
- Docker & Docker Compose
- GitHub OAuth App
- PostgreSQL (or use Neon for free tier)
- Redis (or use Upstash for free tier)

## Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/Ankesh2004/pontoon.git
cd pontoon
cp .env.example .env
```

### 2. Configure Environment

Edit `.env` with your settings:

```bash
# Create GitHub OAuth App at https://github.com/settings/developers
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret

# Generate a strong JWT secret
JWT_SECRET=$(openssl rand -hex 32)

# Set your domain
DEFAULT_DOMAIN=pontoon.yourdomain.com
```

### 3. Local Development

```bash
# Start all services (Postgres, Redis, Traefik, API, Worker)
make dev-local

# Or run services individually:
make run-api      # API server on :8080
make run-worker   # Background worker
```

### 4. Production Deployment

```bash
# Build and start with Docker Compose
make docker-build
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## API Endpoints

### Authentication
- `GET /auth/github` - Redirect to GitHub OAuth
- `GET /auth/callback` - OAuth callback handler

### Projects (Authenticated)
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects` - List user's projects
- `GET /api/v1/projects/:id` - Get project details
- `PUT /api/v1/projects/:id` - Update project
- `DELETE /api/v1/projects/:id` - Delete project

### Deployments (Authenticated)
- `POST /api/v1/projects/:id/deploy` - Trigger deployment
- `GET /api/v1/projects/:id/deployments` - List deployments
- `GET /api/v1/deployments/:id` - Get deployment status
- `GET /api/v1/deployments/:id/logs` - Get runtime logs (REST)
- `GET /api/v1/deployments/:id/ws` - Stream build logs (WebSocket)

### Webhooks
- `POST /webhooks/github` - GitHub webhook endpoint

## Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Build binaries
make build
```

## Deployment Targets

### GCP Free Tier (e2-micro)

1. Create VM instance (e2-micro, 1GB RAM)
2. Install Docker
3. Setup Cloudflare Tunnel
4. Configure Neon PostgreSQL
5. Configure Upstash Redis
6. Deploy with docker-compose

### Memory Budget (1GB VM)

| Component | RAM |
|-----------|-----|
| OS + kernel | ~150MB |
| Go API | ~40MB |
| Go Worker | ~40MB |
| Traefik | ~40MB |
| User containers | ~770MB |
| **Per container limit** | **128MB** |

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `REDIS_URL` | Redis connection string | Yes |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | Yes |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | Yes |
| `JWT_SECRET` | Secret for JWT signing | Yes |
| `DEFAULT_DOMAIN` | Base domain for deployments | Yes |
| `API_ADDR` | API server address | No (default: :8080) |
| `MAX_CONTAINER_MEMORY_MB` | Memory limit per container | No (default: 128) |
| `TOTAL_CONTAINER_MEMORY_MB` | Total memory budget | No (default: 700) |

## Security

- **Tenant Isolation**: Each user gets isolated Docker bridge network
- **Container Hardening**: 
  - Memory/CPU limits via cgroups
  - Drop all Linux capabilities
  - No-new-privileges security option
- **OAuth State**: CSRF protection with state parameter
- **Webhook Verification**: HMAC-SHA256 signature validation

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.
