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

Pontoon is a **100% self-hosted, sovereign PaaS**. No external data
dependencies — PostgreSQL and Redis run inside the same Docker Compose
stack as the API and Worker.

```
┌──────────────────────────────────────────────────────┐
│              Your Server (AWS EC2 / home lab / VM)    │
│                                                       │
│  docker compose up                                    │
│  ┌──────────────────────────────────────────────┐    │
│  │  pontoon-ingress network (shared w/ Traefik) │    │
│  │  ┌───────────┐  ┌──────────┐  ┌───────────┐  │    │
│  │  │ Traefik   │  │ cloudflared│ │ (optional)│  │    │
│  │  │ (routing) │  │ (tunnel)  │  └───────────┘  │    │
│  │  └─────┬─────┘  └─────┬────┘                  │    │
│  │        │              │                        │    │
│  │  ┌─────▼─────────────▼────────────────────┐   │    │
│  │  │  pontoon-api  ◄────────────────────────┼── │    │
│  │  │  pontoon-worker ◄── /var/run/docker.sock│─ │    │
│  │  │             │                            │  │    │
│  │  │             ▼  spawns user containers    │  │    │
│  │  └─────────────┬───────────────────────────┘  │    │
│  │                │  pontoon-<tenant> networks   │    │
│  │   ┌────────────▼───────────────────────────┐  │    │
│  │   │   User Containers (isolated, capped)   │  │    │
│  │   └────────────────────────────────────────┘  │    │
│  └──────────────────────────────────────────────┘    │
│                                                       │
│  pontoon-data network (internal, isolated)            │
│  ┌──────────────┐    ┌──────────────┐                 │
│  │  PostgreSQL  │    │    Redis     │                 │
│  │  (state)     │    │  (queue/pubsub)                 │
│  └──────────────┘    └──────────────┘                 │
└──────────────────────────────────────────────────────┘
                         │
                    Internet / GitHub
```

## Prerequisites

- A capable server (any VM, EC2, or home lab machine; 2GB+ RAM recommended)
- Docker & Docker Compose
- Go 1.23+ (only for local `go run` development)
- A GitHub OAuth App (https://github.com/settings/developers)

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

### 3. Deploy (self-hosted, single command)

The `pontoon-ingress` network must exist before `docker compose up` so
Traefik and the worker-spawned user containers share it. Create it once:

```bash
docker network create pontoon-ingress
```

Then boot the entire stack — Postgres, Redis, Traefik, API, Worker:

```bash
docker compose up -d --build
```

Optional: enable the Cloudflare tunnel (zero-trust ingress, no exposed ports):

```bash
# set CLOUDFLARE_TUNNEL_TOKEN in .env, then:
docker compose --profile tunnel up -d
```

### 4. Local Development (without Docker for the Go binaries)

For iterating on Go code directly on your host, only stand up the data
services and run the binaries with `go run`:

```bash
docker compose -f docker-compose.test.yml up -d   # Postgres + Redis only
make migrate                                       # apply schema
make run-api                                       # API on :8080
make run-worker                                     # background worker
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

## Deployment

Pontoon runs anywhere Docker does. Recommended: a 2GB+ RAM VM
(AWS EC2 t3.small, Hetzner CX22, or a home-lab machine).

1. Install Docker + Docker Compose on the host
2. `git clone` this repo and `cp .env.example .env`
3. Fill in GitHub OAuth credentials and `DEFAULT_DOMAIN`
4. Create the shared ingress network once:
   ```bash
   docker network create pontoon-ingress
   ```
5. Boot the full stack:
   ```bash
   docker compose up -d --build
   ```
6. (Optional) Point a Cloudflare Tunnel at the host and enable the
   `tunnel` profile to avoid exposing any public ports.

### Container Memory Budget

Pontoon no longer targets memory-starved free tiers. Defaults are tuned
for capable servers and are fully configurable via environment:

| Setting | Default | Purpose |
|---------|---------|---------|
| `MAX_CONTAINER_MEMORY_MB` | 256 | Hard cap per deployed user container |
| `TOTAL_CONTAINER_MEMORY_MB` | 2048 | Cluster-wide budget enforced before new builds |

## Configuration

All data services are self-hosted in `docker-compose.yml`. Only secrets
and identity need to be set in `.env` (see `.env.example`).

| Variable | Description | Required |
|----------|-------------|----------|
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | Yes |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | Yes |
| `JWT_SECRET` | Secret for JWT signing | Yes |
| `DEFAULT_DOMAIN` | Base domain for deployments | Yes |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | Self-hosted PG creds | No (defaults: pontoon) |
| `API_ADDR` | API server address | No (default: :8080) |
| `MAX_CONTAINER_MEMORY_MB` | Memory limit per container | No (default: 256) |
| `TOTAL_CONTAINER_MEMORY_MB` | Total memory budget | No (default: 2048) |
| `ACME_EMAIL` | Let's Encrypt email | No |
| `CLOUDFLARE_TUNNEL_TOKEN` | Cloudflare tunnel token | No (tunnel profile) |

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
