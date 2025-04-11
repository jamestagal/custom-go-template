#!/bin/bash

# Navigate to the project root
cd "$(dirname "$0")"

# Run the build command first
echo "Building project..."
./build.sh

# Run the Alpine integration tests
echo "Running Alpine integration tests..."
go test ./tests/alpine/... -v | grep -E "FAIL|ok"

# Check the exit status
status=${PIPESTATUS[0]}
if [ $status -eq 0 ]; then
  echo "All Alpine tests passed!"
  exit 0
else
  echo "Some tests failed. Check the output above for details."
  exit 1
fi
