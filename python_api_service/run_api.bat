@echo off
set VENV_DIR=venv_jm_api
set HOST=0.0.0.0
set PORT=5000
set THREADS=4

if not exist "%VENV_DIR%\Scripts\activate.bat" (
    echo Virtual environment '%VENV_DIR%' not found. Please run install.bat first.
    goto :eof
)

call %VENV_DIR%\Scripts\activate.bat

echo Starting JMComic API server with Waitress on %HOST%:%PORT% with %THREADS% threads...
waitress-serve --listen=%HOST%:%PORT% --threads=%THREADS% api_server:app

rem To stop, press Ctrl+C in the console where waitress-serve is running.
