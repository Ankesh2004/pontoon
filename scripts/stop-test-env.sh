#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Stopping test environment...${NC}"

# Stop API server
if [ -f /tmp/pontoon-api.pid ]; then
    API_PID=$(cat /tmp/pontoon-api.pid)
    kill $API_PID 2>/dev/null || true
    rm /tmp/pontoon-api.pid
    echo -e "${GREEN}✓ API server stopped${NC}"
fi

# Stop Worker
if [ -f /tmp/pontoon-worker.pid ]; then
    WORKER_PID=$(cat /tmp/pontoon-worker.pid)
    kill $WORKER_PID 2>/dev/null || true
    rm /tmp/pontoon-worker.pid
    echo -e "${GREEN}✓ Worker stopped${NC}"
fi

# Stop Docker containers
docker-compose -f docker-compose.test.yml down

echo -e "${GREEN}✓ Test environment stopped${NC}"
