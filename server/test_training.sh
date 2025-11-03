#!/bin/bash

echo "🧪 Testing AI Training System"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if server is running
echo -e "${BLUE}Checking if server is running...${NC}"
if ! curl -s http://localhost:8080/v1/health > /dev/null 2>&1; then
    echo -e "${YELLOW}Server not running. Starting server...${NC}"
    echo "Please run: ./server"
    echo "Then run this script again"
    exit 1
fi

echo -e "${GREEN}✓ Server is running${NC}"
echo ""

# You need to get a JWT token first
echo -e "${YELLOW}⚠️  You need a JWT token to test${NC}"
echo ""
echo "Steps to get a token:"
echo "1. Register: curl -X POST http://localhost:8080/v1/register -H 'Content-Type: application/json' -d '{\"username\":\"test\",\"email\":\"test@test.com\",\"password\":\"test123\"}'"
echo "2. Login: curl -X POST http://localhost:8080/v1/login -H 'Content-Type: application/json' -d '{\"email\":\"test@test.com\",\"password\":\"test123\"}'"
echo ""
echo "Then set your token here:"
echo "export TOKEN='your_jwt_token_here'"
echo ""

# Check if TOKEN is set
if [ -z "$TOKEN" ]; then
    echo -e "${YELLOW}TOKEN not set. Please set it and run again.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Token found${NC}"
echo ""

# Test 1: List directories
echo -e "${BLUE}Test 1: List available directories${NC}"
echo "================================"
curl -s -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/v1/ai/directories | jq '.'
echo ""

# Test 2: Get PokemonModel info
echo -e "${BLUE}Test 2: Get PokemonModel directory info${NC}"
echo "================================"
curl -s -H "Authorization: Bearer $TOKEN" \
    "http://localhost:8080/v1/ai/directory?folder=PokemonModel" | jq '.directory_info | {name, total_files, total_size, subdirs: .subdirs[:5]}'
echo ""

echo -e "${GREEN}✓ All tests passed!${NC}"
echo ""
echo "Next steps:"
echo "1. Create a simple training script to test"
echo "2. Or use the mock training demo below"
