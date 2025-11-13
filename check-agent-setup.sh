#!/usr/bin/env bash

echo "============================================================"
echo "🔍 AI Training Agent Setup Checker"
echo "============================================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if server is running
echo "1️⃣  Checking if server is running..."
if lsof -i :8081 > /dev/null 2>&1; then
    echo -e "   ${GREEN}✅ Server is running on port 8081${NC}"
else
    echo -e "   ${RED}❌ Server is NOT running on port 8081${NC}"
    echo "   To start the server:"
    echo "   cd server && go run cmd/main.go"
    echo ""
fi

# Check if database is accessible
echo "2️⃣  Checking database connection..."
if psql -h localhost -U postgres -d aimanage -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "   ${GREEN}✅ Database is accessible${NC}"
else
    echo -e "   ${RED}❌ Cannot connect to database${NC}"
    echo "   Make sure PostgreSQL is running and the 'aimanage' database exists"
    echo ""
fi

# Check if users have API keys
echo "3️⃣  Checking user API keys..."
API_KEY_COUNT=$(psql -h localhost -U postgres -d aimanage -t -c "SELECT COUNT(*) FROM users WHERE api_key IS NOT NULL AND api_key != ''" 2>/dev/null | tr -d ' ')

if [ -z "$API_KEY_COUNT" ]; then
    echo -e "   ${RED}❌ Could not query database${NC}"
elif [ "$API_KEY_COUNT" -gt 0 ]; then
    echo -e "   ${GREEN}✅ Found $API_KEY_COUNT user(s) with API keys${NC}"

    # Show API keys (first 12 chars only)
    echo "   Your API key(s):"
    psql -h localhost -U postgres -d aimanage -t -c "SELECT email, SUBSTRING(api_key, 1, 12) || '...' as api_key_preview FROM users WHERE api_key IS NOT NULL" 2>/dev/null | grep -v "^$"
else
    echo -e "   ${YELLOW}⚠️  No users have API keys${NC}"
    echo "   Generating API keys for all users..."
    psql -h localhost -U postgres -d aimanage -c "UPDATE users SET api_key = 'sk_live_' || substr(md5(random()::text || email), 1, 24) WHERE api_key IS NULL" 2>/dev/null
    echo -e "   ${GREEN}✅ API keys generated${NC}"
fi

# Check if migrations have been run
echo "4️⃣  Checking if api_key column exists..."
HAS_API_KEY=$(psql -h localhost -U postgres -d aimanage -t -c "SELECT column_name FROM information_schema.columns WHERE table_name='users' AND column_name='api_key'" 2>/dev/null | tr -d ' ')

if [ "$HAS_API_KEY" == "api_key" ]; then
    echo -e "   ${GREEN}✅ api_key column exists${NC}"
else
    echo -e "   ${RED}❌ api_key column does NOT exist${NC}"
    echo "   Run migrations:"
    echo "   cd server && make migrate-up"
    echo ""
fi

# Check if Python dependencies are installed
echo "5️⃣  Checking Python dependencies..."
if python3 -c "import websockets, torch" 2>/dev/null; then
    echo -e "   ${GREEN}✅ Python dependencies installed${NC}"
else
    echo -e "   ${YELLOW}⚠️  Missing Python dependencies${NC}"
    echo "   Install with: pip3 install websockets torch"
    echo ""
fi

# Check if training agent script exists
echo "6️⃣  Checking training agent script..."
if [ -f "training-agent/train_agent.py" ]; then
    echo -e "   ${GREEN}✅ Training agent script found${NC}"
else
    echo -e "   ${RED}❌ Training agent script not found${NC}"
    echo ""
fi

echo ""
echo "============================================================"
echo "📋 Summary"
echo "============================================================"

# Get a user's API key to show in the command
USER_EMAIL=$(psql -h localhost -U postgres -d aimanage -t -c "SELECT email FROM users LIMIT 1" 2>/dev/null | tr -d ' ')
USER_API_KEY=$(psql -h localhost -U postgres -d aimanage -t -c "SELECT api_key FROM users WHERE email='$USER_EMAIL'" 2>/dev/null | tr -d ' ')

if [ ! -z "$USER_API_KEY" ] && [ "$USER_API_KEY" != "" ]; then
    echo "To connect your training agent, run:"
    echo ""
    echo "cd training-agent"
    echo "python3 train_agent.py --api-key \"$USER_API_KEY\""
    echo ""
else
    echo -e "${YELLOW}⚠️  Cannot retrieve API key. Make sure to:"
    echo "1. Start the server"
    echo "2. Log in to the frontend"
    echo "3. Get your API key from the Settings page"
    echo "4. Run: python3 training-agent/train_agent.py --api-key YOUR_API_KEY${NC}"
    echo ""
fi

echo "============================================================"
