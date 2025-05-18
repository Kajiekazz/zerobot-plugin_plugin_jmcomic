@echo off
set PYTHON_CMD=python
set VENV_DIR=venv_jm_api

echo Checking for Python...
%PYTHON_CMD% --version > nul 2>&1
if errorlevel 1 (
    echo %PYTHON_CMD% could not be found. Please install Python and add it to PATH.
    goto :eof
)

echo Checking for pip...
%PYTHON_CMD% -m pip --version > nul 2>&1
if errorlevel 1 (
    echo pip could not be found for %PYTHON_CMD%. Please ensure pip is installed.
    goto :eof
)

echo Creating virtual environment in %VENV_DIR%...
%PYTHON_CMD% -m venv %VENV_DIR%
if errorlevel 1 (
    echo Failed to create virtual environment.
    goto :eof
)

echo Activating virtual environment and installing dependencies...
call %VENV_DIR%\Scripts\activate.bat
pip install -r requirements.txt
if errorlevel 1 (
    echo Failed to install dependencies.
    goto :eof
)

echo.
echo Setup complete!
echo To activate the virtual environment manually, run: %VENV_DIR%\Scripts\activate.bat
echo To run the API server (example for development): python api_server.py
echo For production, consider using Waitress: waitress-serve --listen=0.0.0.0:5000 api_server:app
echo Make sure you have configured python_api_service/jm.yaml correctly.

rem Deactivate after script finishes (optional)
rem call %VENV_DIR%\Scripts\deactivate.bat
