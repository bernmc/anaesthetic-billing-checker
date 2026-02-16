#!/bin/bash
# stop-billing-app.command
# Double-click this file in Finder to stop the Anaesthetic Billing Checker.
#
# Copyright (c) 2026 Bernard McClement
# Licensed under the MIT License. See LICENSE file in the project root.

echo "Stopping Anaesthetic Billing Checker on port 8080…"

PID=$(lsof -ti :8080 2>/dev/null)
if [ -n "$PID" ]; then
  kill "$PID"
  echo "Stopped (PID $PID)."
else
  echo "No process found on port 8080."
fi

echo ""
echo "Done. You can close this window."
sleep 2
