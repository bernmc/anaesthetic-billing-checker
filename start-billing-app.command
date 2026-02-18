#!/bin/bash
# start-billing-app.command
# Double-click this file in Finder to start the Anaesthetic Billing Checker.
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
  echo "Press Ctrl+C or double-click stop-billing-app.command to stop."
  echo ""
  ./billing-checker
else
  # Node.js fallback for development
  if ! command -v node &>/dev/null; then
    echo "Error: No billing-checker binary found and Node.js is not installed."
    echo "Download the standalone binary from the GitHub Releases page,"
    echo "or install Node.js 18+ from https://nodejs.org"
    echo ""
    read -n 1 -s -r -p "Press any key to close…"
    exit 1
  fi
  if [ ! -d "node_modules" ]; then
    echo "First run detected — installing dependencies…"
    npm install
    echo ""
  fi
  echo "Starting server (Node.js)…"
  echo "The app will open at http://localhost:8080"
  echo "Press Ctrl+C or double-click stop-billing-app.command to stop."
  echo ""
  (sleep 2 && open "http://localhost:8080") &
  npm start
fi
