#!/bin/bash
VENV_DIR="venv_jm_api"
HOST="0.0.0.0"
PORT="5000" # 和 api_server.py 以及 Go 插件配置中保持一致
WORKERS="4" # Gunicorn worker 数量，通常是 (2 * CPU核心数) + 1

# 检查虚拟环境是否存在
if [ ! -d "$VENV_DIR" ]; then
    echo "Virtual environment '$VENV_DIR' not found. Please run install.sh first."
    exit 1
fi

source $VENV_DIR/bin/activate

echo "Starting JMComic API server with Gunicorn on $HOST:$PORT with $WORKERS workers..."
# Gunicorn 默认会以后台模式运行，除非指定 --daemon
# 如果要看日志，可以不加 --daemon，或者配置Gunicorn的日志输出
# gunicorn --workers $WORKERS --bind $HOST:$PORT api_server:app --log-level info
gunicorn --workers $WORKERS --bind $HOST:$PORT api_server:app
# 要停止，需要找到gunicorn进程并kill

deactivate
