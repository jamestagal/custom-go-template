#!/bin/bash
cd "/Users/benjaminwaller/Projects/Jim Fisk/custom_go_template"
./server alpine_integration_test | grep -A 5 FAIL | head -n 20
