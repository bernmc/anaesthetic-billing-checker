#!/bin/bash
# start-billing-app.sh
# Run this script to start the Anaesthetic Billing Checker.
# Prefers the standalone binary; falls back to Node.js if not found.
#
# Copyright (c) 2026 Bernard McClement
# Licensed under the MIT License. See LICENSE file in the project root.

cd "$(dirname "$0")"

echo "=== Anaesthetic Billing Checker ==="
echo ""

if [ -x "./billing-checker" ]; then
  echo "Starting standalone server…"
  echo "The app will open at http://localhost:8080"
  echo "Press Ctrl+C to stop."
  echo ""
  ./billing-checker
else
  if ! command -v node &>/dev/null; then
    echo "Error: No billing-checker binary found and Node.js is not installed."
    echo "Download the standalone binary from the GitHub Releases page,"
    echo "or install Node.js 18+ from https://nodejs.org"
    exit 1
  fi
  if [ ! -d "node_modules" ]; then
    echo "First run detected — installing dependencies…"
    npm install
    echo ""
  fi
  echo "Starting server (Node.js)…"
  echo "The app will open at http://localhost:8080"
  echo "Press Ctrl+C to stop."
  echo ""
  (sleep 2 && xdg-open "http://localhost:8080" 2>/dev/null) &
  npm start
fi
