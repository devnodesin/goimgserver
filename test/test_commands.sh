#!/bin/bash

# Test script for command endpoints
# This script demonstrates the command endpoints functionality

# Ensure SERVER_URL is used consistently
SERVER_URL="${SERVER_URL:-http://localhost:9000}"

# Update all curl commands to use SERVER_URL
response=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER_URL/ping" 2>/dev/null)
if [ "$response" -ne 200 ]; then
  echo "Ping test failed! Server response code: $response"
else
  echo "Ping test succeeded!"
fi

response=$(curl -s -X POST "$SERVER_URL/cmd/clear" 2>/dev/null)
if [ "$response" != "null" ]; then
  echo "Clear command test failed! Unexpected response: $response"
else
  echo "Clear command test succeeded!"
fi

response=$(curl -s -X POST "$SERVER_URL/cmd/gitupdate" 2>/dev/null)
if [ "$response" != "null" ]; then
  echo "Git update command test failed! Unexpected response: $response"
else
  echo "Git update command test succeeded!"
fi

response=$(curl -s -X POST "$SERVER_URL/cmd/invalid" 2>/dev/null)
if [ "$response" != "null" ]; then
  echo "Invalid command test failed! Unexpected response: $response"
else
  echo "Invalid command test succeeded!"
fi