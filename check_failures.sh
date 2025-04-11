#!/bin/bash

# Run the alpine tests and check for failures
cd "$(dirname "$0")"
./run_tests.sh tests/alpine | grep -E "FAIL|✖"

# If nothing was found, tests passed
if [ ${PIPESTATUS[1]} -ne 0 ]; then
  echo "All tests appear to have passed!"
  exit 0
else
  echo "Some tests are still failing."
  exit 1
fi
