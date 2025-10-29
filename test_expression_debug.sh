#!/bin/bash

# Test script for expression debugging
# Usage: ./test_expression_debug.sh

echo "=== Expression Debugging Test ==="
echo ""
echo "Testing with DEBUG_EXPRESSIONS=true..."
echo ""

# Run server with debugging enabled
DEBUG_EXPRESSIONS=true go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo ""
echo "Server started with PID $SERVER_PID"
echo "Visit http://localhost:3333 to see expression debug logs"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Wait for user interrupt
wait $SERVER_PID
