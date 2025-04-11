#!/bin/bash
# Find the Go executable
GO_PATH=$(which go)

if [ -z "$GO_PATH" ]; then
    echo "Go not found in PATH. Looking for common installation paths..."
    
    # Check common Go installation paths
    POTENTIAL_PATHS=(
        "/usr/local/go/bin/go"
        "/usr/local/bin/go"
        "/opt/homebrew/bin/go"
        "/snap/bin/go"
        "$HOME/go/bin/go"
        "$HOME/.go/bin/go"
    )
    
    for path in "${POTENTIAL_PATHS[@]}"; do
        if [ -x "$path" ]; then
            GO_PATH="$path"
            echo "Found Go at $GO_PATH"
            break
        fi
    done
    
    if [ -z "$GO_PATH" ]; then
        echo "Go not found. Please install Go or add it to your PATH."
        exit 1
    fi
fi

echo "Using Go at $GO_PATH"
echo "Building project..."
"$GO_PATH" build ./...
