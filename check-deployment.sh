#!/bin/bash
# Deployment Status Check Script

echo "=================================="
echo "AiManage Deployment Status"
echo "=================================="
echo ""

echo "1️⃣  Docker Containers:"
echo "----------------------------------"
docker compose -f /opt/aimanage/docker-compose.prod.yml ps
echo ""

echo "2️⃣  Container Health:"
echo "----------------------------------"
RUNNING=$(docker compose -f /opt/aimanage/docker-compose.prod.yml ps -q | wc -l)
echo "Running containers: $RUNNING/3"
if [ "$RUNNING" -eq 3 ]; then
    echo "✅ All containers are running!"
else
    echo "⚠️  Some containers are not running"
fi
echo ""

echo "3️⃣  Recent Logs (last 10 lines per service):"
echo "----------------------------------"
docker compose -f /opt/aimanage/docker-compose.prod.yml logs --tail=10
echo ""

echo "4️⃣  Application URLs:"
echo "----------------------------------"
echo "Frontend:  http://109.199.115.1:3000"
echo "Backend:   http://109.199.115.1:8081"
echo "Database:  localhost:5432 (internal)"
echo ""

echo "=================================="
echo "To view live logs: docker compose -f /opt/aimanage/docker-compose.prod.yml logs -f"
echo "To restart: docker compose -f /opt/aimanage/docker-compose.prod.yml restart"
echo "=================================="
