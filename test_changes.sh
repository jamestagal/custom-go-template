#!/bin/bash

# Navigate to the project root
cd "$(dirname "$0")"

# Run all relevant tests
echo "Running Alpine integration tests..."
go test -v ./tests/alpine

echo "Running transformer tests..."
go test -v ./transformer

# Check the exit status
if [ $? -eq 0 ]; then
  echo "All tests passed!"
  exit 0
else
  echo "Some tests failed. Check the output above for details."
  exit 1
fi
