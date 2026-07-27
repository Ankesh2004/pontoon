#!/bin/bash

# Exit on error
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting local test environment...${NC}"

# Stop any existing containers
docker-compose -f docker-compose.test.yml down -v

# Start PostgreSQL and Redis
docker-compose -f docker-compose.test.yml up -d

# Wait for PostgreSQL to be ready
echo -e "${YELLOW}Waiting for PostgreSQL...${NC}"
sleep 5

# Wait for Redis to be ready
echo -e "${YELLOW}Waiting for Redis...${NC}"
sleep 3

# Check if services are running
if ! docker-compose -f docker-compose.test.yml ps | grep -q "Up"; then
    echo -e "${RED}Failed to start services${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Services started successfully${NC}"

# Create .env file for testing
cat > .env << EOF
DATABASE_URL=postgresql://pontoon:pontoon@localhost:5432/pontoon?sslmode=disable
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=test-secret-key-for-local-testing
GITHUB_CLIENT_ID=test-client-id
GITHUB_CLIENT_SECRET=test-client-secret
API_ADDR=:8080
DEFAULT_DOMAIN=localhost
EOF

echo -e "${GREEN}✓ Environment configured${NC}"

# Run migrations
echo -e "${YELLOW}Running database migrations...${NC}"
go run cmd/migrate/main.go

echo -e "${GREEN}✓ Migrations completed${NC}"

# Start API server in background
echo -e "${YELLOW}Starting API server...${NC}"
go run cmd/api/main.go &
API_PID=$!

# Wait for API to be ready
sleep 3

# Check if API is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${RED}API server failed to start${NC}"
    kill $API_PID 2>/dev/null || true
    exit 1
fi

echo -e "${GREEN}✓ API server started on port 8080${NC}"

# Start Worker in background
echo -e "${YELLOW}Starting Worker...${NC}"
go run cmd/worker/main.go &
WORKER_PID=$!

sleep 2
echo -e "${GREEN}✓ Worker started${NC}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Local test environment is ready!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "API Server: http://localhost:8080"
echo "Health Check: curl http://localhost:8080/health"
echo ""
echo "To test endpoints, use the test-api.sh script:"
echo "  ./scripts/test-api.sh"
echo ""
echo "To stop the environment:"
echo "  ./scripts/stop-test-env.sh"
echo ""

# Save PIDs for cleanup
echo $API_PID > /tmp/pontoon-api.pid
echo $WORKER_PID > /tmp/pontoon-worker.pid
