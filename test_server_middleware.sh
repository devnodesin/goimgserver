#!/bin/bash

# Test script for server middleware functionality
# This script demonstrates the enhanced server features

echo "================================"
echo "Server Middleware Test Script"
echo "================================"
echo ""

# Note: This script requires a running server instance
SERVER_URL="${SERVER_URL:-http://localhost:9000}"

echo "Testing server at: $SERVER_URL"
echo ""

# Test 1: Health endpoint
echo "1. Testing /health endpoint..."
response=$(curl -s "$SERVER_URL/health" 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "   ✅ Health endpoint responding"
    echo "   Response: $(echo $response | head -c 100)..."
else
    echo "   ❌ Server not running or health endpoint not accessible"
    echo "   Please start the server first"
    exit 1
fi
echo ""

# Test 2: Liveness probe
echo "2. Testing /live endpoint..."
response=$(curl -s "$SERVER_URL/live" 2>/dev/null)
if echo "$response" | grep -q "alive"; then
    echo "   ✅ Liveness probe working"
else
    echo "   ⚠️  Liveness probe not responding correctly"
fi
echo ""

# Test 3: Readiness probe
echo "3. Testing /ready endpoint..."
response=$(curl -s "$SERVER_URL/ready" 2>/dev/null)
if echo "$response" | grep -q "ready"; then
    echo "   ✅ Readiness probe working"
else
    echo "   ⚠️  Readiness probe not responding correctly"
fi
echo ""

# Test 4: Request ID middleware
echo "4. Testing Request ID middleware..."
headers=$(curl -s -I "$SERVER_URL/ping" 2>/dev/null)
if echo "$headers" | grep -qi "X-Request-ID"; then
    request_id=$(echo "$headers" | grep -i "X-Request-ID" | cut -d: -f2 | tr -d ' \r')
    echo "   ✅ Request ID middleware working"
    echo "   Request ID: $request_id"
else
    echo "   ⚠️  Request ID not found in headers"
fi
echo ""

# Test 5: Security headers
echo "5. Testing Security headers..."
headers=$(curl -s -I "$SERVER_URL/ping" 2>/dev/null)
security_headers=("X-Content-Type-Options" "X-Frame-Options" "X-XSS-Protection" "Referrer-Policy")
for header in "${security_headers[@]}"; do
    if echo "$headers" | grep -qi "$header"; then
        echo "   ✅ $header present"
    else
        echo "   ⚠️  $header missing"
    fi
done
echo ""

# Test 6: CORS headers
echo "6. Testing CORS headers..."
headers=$(curl -s -I -H "Origin: http://example.com" "$SERVER_URL/ping" 2>/dev/null)
if echo "$headers" | grep -qi "Access-Control-Allow-Origin"; then
    echo "   ✅ CORS headers present"
else
    echo "   ⚠️  CORS headers not found"
fi
echo ""

# Test 7: Error handling
echo "7. Testing error handling..."
response=$(curl -s "$SERVER_URL/nonexistent" 2>/dev/null)
if echo "$response" | grep -q "error"; then
    echo "   ✅ Error handling working (returns error for 404)"
else
    echo "   ⚠️  Error handling might not be working correctly"
fi
echo ""

# Test 8: Rate limiting (if enabled)
echo "8. Testing rate limiting (if enabled)..."
echo "   Making 10 rapid requests..."
rate_limited=0
for i in {1..10}; do
    status=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER_URL/ping" 2>/dev/null)
    if [ "$status" = "429" ]; then
        rate_limited=$((rate_limited + 1))
    fi
done

if [ $rate_limited -gt 0 ]; then
    echo "   ✅ Rate limiting working ($rate_limited requests limited)"
else
    echo "   ℹ️  Rate limiting not triggered (may be disabled or limit not reached)"
fi
echo ""

echo "================================"
echo "Test Summary"
echo "================================"
echo "All middleware features have been tested."
echo "Check the results above for any issues."
echo ""
echo "Note: Some tests may show warnings if the server"
echo "is not running or middleware is not fully configured."
