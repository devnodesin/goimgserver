#!/bin/bash

# Test script for server middleware functionality
# This script demonstrates the enhanced server features

# Ensure SERVER_URL is used consistently
SERVER_URL="${SERVER_URL:-http://localhost:9000}"

# Update all curl commands to use SERVER_URL
response=$(curl -s "$SERVER_URL/health" 2>/dev/null)
echo "Health check response: $response"

response=$(curl -s "$SERVER_URL/live" 2>/dev/null)
echo "Live check response: $response"

response=$(curl -s "$SERVER_URL/ready" 2>/dev/null)
echo "Ready check response: $response"

headers=$(curl -s -I "$SERVER_URL/ping" 2>/dev/null)
echo "Ping headers: $headers"

headers=$(curl -s -I -H "Origin: http://example.com" "$SERVER_URL/ping" 2>/dev/null)
echo "Ping headers with origin: $headers"

response=$(curl -s "$SERVER_URL/nonexistent" 2>/dev/null)
echo "Non-existent endpoint response: $response"

status=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER_URL/ping" 2>/dev/null)
echo "Ping status code: $status"