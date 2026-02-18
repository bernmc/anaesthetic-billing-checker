@echo off
rem start-billing-app.bat
rem Double-click this file to start the Anaesthetic Billing Checker.
rem Prefers the standalone binary; falls back to Node.js if not found.
rem
rem Copyright (c) 2026 Bernard McClement
rem Licensed under the MIT License. See LICENSE file in the project root.

cd /d "%~dp0"

echo === Anaesthetic Billing Checker ===
echo.

if exist "billing-checker.exe" (
  echo Starting standalone server...
  echo The app will open at http://localhost:8080
  echo Press Ctrl+C to stop.
  echo.
  billing-checker.exe
) else (
  where node >nul 2>&1
  if errorlevel 1 (
    echo Error: No billing-checker.exe found and Node.js is not installed.
    echo Download the standalone binary from the GitHub Releases page,
    echo or install Node.js 18+ from https://nodejs.org
    echo.
    pause
    exit /b 1
  )
  if not exist "node_modules" (
    echo First run detected - installing dependencies...
    call npm install
    echo.
  )
  echo Starting server (Node.js)...
  echo The app will open at http://localhost:8080
  echo Press Ctrl+C to stop.
  echo.
  start "" /b cmd /c "timeout /t 2 /nobreak >nul & start http://localhost:8080"
  call npm start
)
