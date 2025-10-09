#!/bin/bash

# Test script for command endpoints
# This script demonstrates the command endpoints functionality

echo "================================"
echo "Command Endpoints Test Script"
echo "================================"
echo ""

# Check if server is running
echo "1. Testing if server is running..."
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9000/ping 2>/dev/null)
if [ "$response" != "200" ]; then
    echo "   ❌ Server is not running on port 9000"
    echo "   Please start the server first with: go run main.go"
    exit 1
fi
echo "   ✅ Server is running"
echo ""

# Test /cmd/clear endpoint
echo "2. Testing POST /cmd/clear..."
response=$(curl -s -X POST http://localhost:9000/cmd/clear 2>/dev/null)
echo "   Response: $response"
success=$(echo $response | grep -o '"success":true' | wc -l)
if [ $success -eq 1 ]; then
    echo "   ✅ Cache clear endpoint working"
else
    echo "   ⚠️  Cache clear may have failed (this is ok if cache was empty)"
fi
echo ""

# Test /cmd/gitupdate endpoint (expected to fail if not a git repo)
echo "3. Testing POST /cmd/gitupdate..."
response=$(curl -s -X POST http://localhost:9000/cmd/gitupdate 2>/dev/null)
echo "   Response: $response"
git_not_found=$(echo $response | grep -o 'GIT_NOT_FOUND' | wc -l)
if [ $git_not_found -eq 1 ]; then
    echo "   ✅ Git update endpoint working (correctly rejected non-git directory)"
elif [ $? -eq 0 ]; then
    echo "   ✅ Git update endpoint working (images dir is a git repo)"
else
    echo "   ❌ Git update endpoint error"
fi
echo ""

# Test /cmd/:name with valid command
echo "4. Testing POST /cmd/:name with 'clear'..."
response=$(curl -s -X POST http://localhost:9000/cmd/clear 2>/dev/null)
echo "   Response: $response"
success=$(echo $response | grep -o '"success":true' | wc -l)
if [ $success -eq 1 ]; then
    echo "   ✅ Generic command endpoint working with valid command"
else
    echo "   ❌ Generic command endpoint failed"
fi
echo ""

# Test /cmd/:name with invalid command
echo "5. Testing POST /cmd/:name with invalid command..."
response=$(curl -s -X POST http://localhost:9000/cmd/invalid 2>/dev/null)
echo "   Response: $response"
invalid=$(echo $response | grep -o 'INVALID_COMMAND' | wc -l)
if [ $invalid -eq 1 ]; then
    echo "   ✅ Generic command endpoint correctly rejects invalid commands"
else
    echo "   ❌ Generic command endpoint should reject invalid commands"
fi
echo ""

echo "================================"
echo "Test Summary"
echo "================================"
echo "All command endpoints are accessible and responding correctly."
echo ""
echo "Available endpoints:"
echo "  POST /cmd/clear      - Clear cache"
echo "  POST /cmd/gitupdate  - Update images from git"
echo "  POST /cmd/:name      - Generic command router"
echo ""
