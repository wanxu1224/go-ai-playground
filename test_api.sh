#!/bin/bash
echo "🧪 Testing Go Weather API..."
echo ""
echo "1. Health Check:"
curl -s http://localhost:8080/health | jq .
echo ""
echo "2. Fetch Beijing Weather (Real-time):"
curl -s "http://localhost:8080/api/fetch?city=beijing" | jq .
echo ""
echo "3. Fetch Multi-City (Concurrent):"
curl -s http://localhost:8080/api/multi-fetch | jq .
echo ""
echo "4. History Records:"
curl -s "http://localhost:8080/history?limit=5" | jq .
