#!/bin/bash

# Custom Go Template Server Runner
# Usage:
#   ./server.sh          - Run normally (debug OFF)
#   ./server.sh debug    - Run with expression debugging (debug ON)

cd "$(dirname "$0")"

if [ "$1" == "debug" ]; then
    echo "🔍 Starting server with DEBUG_EXPRESSIONS=true"
    echo "   You'll see [EXPR-DEBUG] logs showing build-time vs runtime decisions"
    echo ""
    DEBUG_EXPRESSIONS=true go run cmd/server/main.go
else
    echo "🚀 Starting server in normal mode (no debugging)"
    echo "   To enable debugging, run: ./server.sh debug"
    echo ""
    go run cmd/server/main.go
fi
