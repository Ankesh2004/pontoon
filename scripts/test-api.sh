#!/bin/bash

# API Test Script for Pontoon
# This script tests all API endpoints

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        echo "  Response: $3"
        ((TESTS_FAILED++))
    fi
}

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Pontoon API Test Suite${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Test 1: Health check
echo "Test 1: Health Check"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
if [ "$RESPONSE" = "200" ]; then
    print_result 0 "Health endpoint returns 200"
else
    print_result 1 "Health endpoint returns 200" "Got $RESPONSE"
fi
echo ""

# Test 2: Auth endpoint (should redirect to GitHub)
echo "Test 2: GitHub OAuth Redirect"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/github")
if [ "$RESPONSE" = "302" ] || [ "$RESPONSE" = "307" ]; then
    print_result 0 "Auth endpoint redirects to GitHub"
else
    print_result 1 "Auth endpoint redirects to GitHub" "Got $RESPONSE"
fi
echo ""

# Test 3: Protected endpoint without token (should fail)
echo "Test 3: Protected Endpoint Without Token"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/projects")
if [ "$RESPONSE" = "401" ]; then
    print_result 0 "Protected endpoint returns 401 without token"
else
    print_result 1 "Protected endpoint returns 401 without token" "Got $RESPONSE"
fi
echo ""

# Test 4: Webhook endpoint without signature (should fail)
echo "Test 4: Webhook Without Signature"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/webhooks/github")
if [ "$RESPONSE" = "400" ] || [ "$RESPONSE" = "401" ]; then
    print_result 0 "Webhook endpoint rejects requests without signature"
else
    print_result 1 "Webhook endpoint rejects requests without signature" "Got $RESPONSE"
fi
echo ""

# Test 5: Webhook with invalid signature
echo "Test 5: Webhook With Invalid Signature"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "X-Hub-Signature-256: sha256=invalid" \
    -H "Content-Type: application/json" \
    -d '{"action":"push"}' \
    "http://localhost:8080/webhooks/github?project_id=test-project-id")
if [ "$RESPONSE" = "401" ] || [ "$RESPONSE" = "500" ]; then
    print_result 0 "Webhook endpoint rejects invalid signature"
else
    print_result 1 "Webhook endpoint rejects invalid signature" "Got $RESPONSE"
fi
echo ""

# Test 6: Create project without auth (should fail)
echo "Test 6: Create Project Without Auth"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"name":"test-project","repo_url":"https://github.com/test/repo","branch":"main"}' \
    "$BASE_URL/api/v1/projects")
if [ "$RESPONSE" = "401" ]; then
    print_result 0 "Create project requires authentication"
else
    print_result 1 "Create project requires authentication" "Got $RESPONSE"
fi
echo ""

# Test 7: Get deployment without auth (should fail)
echo "Test 7: Get Deployment Without Auth"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/deployments/test-id")
if [ "$RESPONSE" = "401" ]; then
    print_result 0 "Get deployment requires authentication"
else
    print_result 1 "Get deployment requires authentication" "Got $RESPONSE"
fi
echo ""

# Test 8: WebSocket endpoint without auth
echo "Test 8: WebSocket Without Auth"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/ws/logs")
if [ "$RESPONSE" = "401" ] || [ "$RESPONSE" = "400" ]; then
    print_result 0 "WebSocket endpoint requires authentication"
else
    print_result 1 "WebSocket endpoint requires authentication" "Got $RESPONSE"
fi
echo ""

# Summary
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi
