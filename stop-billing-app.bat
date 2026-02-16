@echo off
rem stop-billing-app.bat
rem Double-click this file to stop the Anaesthetic Billing Checker.
rem
rem Copyright (c) 2026 Bernard McClement
rem Licensed under the MIT License. See LICENSE file in the project root.

echo Stopping Anaesthetic Billing Checker on port 8080...

for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080 " ^| findstr "LISTENING"') do (
  taskkill /PID %%a /F >nul 2>&1
  echo Stopped (PID %%a).
)

echo.
echo Done. You can close this window.
timeout /t 3 /nobreak >nul
