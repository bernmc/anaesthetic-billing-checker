#!/bin/bash
# start-billing-app.sh
# Run this script to start the Anaesthetic Billing Checker.
#
# Copyright (c) 2026 Bernard McClement
# Licensed under the MIT License. See LICENSE file in the project root.

cd "$(dirname "$0")"

echo "=== Anaesthetic Billing Checker ==="
echo ""

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
  echo "First run detected — installing dependencies…"
  npm install
  echo ""
fi

echo "Starting server…"
echo "The app will open at http://localhost:8080"
echo "Press Ctrl+C to stop."
echo ""

# Open browser after a short delay
(sleep 2 && xdg-open "http://localhost:8080" 2>/dev/null) &

npm start
