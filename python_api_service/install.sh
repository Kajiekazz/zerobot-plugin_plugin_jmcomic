#!/bin/bash

PYTHON_CMD="python3" # 或者 python, 根据你的环境
VENV_DIR="venv_jm_api"

echo "Checking for Python..."
if ! command -v $PYTHON_CMD &> /dev/null
then
    echo "$PYTHON_CMD could not be found. Please install Python 3."
    exit 1
fi

echo "Checking for pip..."
if ! $PYTHON_CMD -m pip --version &> /dev/null
then
    echo "pip could not be found for $PYTHON_CMD. Please ensure pip is installed."
    exit 1
fi

echo "Creating virtual environment in $VENV_DIR..."
$PYTHON_CMD -m venv $VENV_DIR
if [ $? -ne 0 ]; then
    echo "Failed to create virtual environment."
    exit 1
fi

echo "Activating virtual environment..."
source $VENV_DIR/bin/activate

echo "Installing dependencies from requirements.txt..."
pip install -r requirements.txt
if [ $? -ne 0 ]; then
    echo "Failed to install dependencies."
    # Deactivate on failure if you want
    # deactivate
    exit 1
fi

echo ""
echo "Setup complete!"
echo "To activate the virtual environment manually, run: source $VENV_DIR/bin/activate"
echo "To run the API server (example for development): python api_server.py"
echo "For production, consider using Gunicorn: gunicorn -w 4 -b 0.0.0.0:5000 api_server:app"
echo "Make sure you have configured python_api_service/jm.yaml correctly."

# Deactivate after script finishes (optional, user can activate manually)
# deactivate
