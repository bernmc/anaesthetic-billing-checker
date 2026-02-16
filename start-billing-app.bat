@echo off
rem start-billing-app.bat
rem Double-click this file to start the Anaesthetic Billing Checker.
rem
rem Copyright (c) 2026 Bernard McClement
rem Licensed under the MIT License. See LICENSE file in the project root.

cd /d "%~dp0"

echo === Anaesthetic Billing Checker ===
echo.

rem Install dependencies if needed
if not exist "node_modules" (
  echo First run detected - installing dependencies...
  call npm install
  echo.
)

echo Starting server...
echo The app will open at http://localhost:8080
echo Press Ctrl+C to stop.
echo.

rem Open browser after a short delay
start "" /b cmd /c "timeout /t 2 /nobreak >nul & start http://localhost:8080"

call npm start
