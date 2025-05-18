
## 安装与部署

部署此插件需要两个主要步骤：
1.  部署Python API服务。
2.  编译和安装Go ZeroBot插件。

### 步骤 1: 部署 Python API 服务

Python API服务负责与JMComic网站交互。

**先决条件:**
-   Python 3.7+
-   pip

**安装:**

1.  进入 `python_api_service` 目录:
    ```bash
    cd python_api_service
    ```

2.  **重要**: 配置 `jm.yaml` 文件。
    请参考 `jm.yaml.example` (来自 [JMComic-Crawler-Python](https://github.com/hect0x7/JMComic-Crawler-Python/blob/main/assets/jm.yaml.example)) 创建并配置你自己的 `jm.yaml`。你需要设置正确的JM域名、可能的代理、以及 **`option.dir.base_dir`** (API服务下载漫画的根目录)。

3.  运行安装脚本:
    -   **Linux/macOS**:
        ```bash
        chmod +x install.sh
        ./install.sh
        ```
    -   **Windows**:
        ```batch
        install.bat
        ```
    此脚本会创建一个Python虚拟环境 (`venv_jm_api`) 并安装所需的依赖。

**运行API服务:**

1.  激活虚拟环境:
    -   **Linux/macOS**: `source venv_jm_api/bin/activate`
    -   **Windows**: `venv_jm_api\Scripts\activate.bat`

2.  运行API服务器:
    -   **开发/测试 (使用Flask内置服务器):**
        ```bash
        python api_server.py
        ```
        默认监听 `0.0.0.0:5000`。你可以通过环境变量 `API_HOST` 和 `API_PORT` 修改。
    -   **生产环境:**
        -   **Linux/macOS (使用 Gunicorn):**
            ```bash
            # 示例: 监听 0.0.0.0:5000，使用4个worker进程
            gunicorn -w 4 -b 0.0.0.0:5000 api_server:app
            # 或者使用提供的 run_api.sh (可能需要调整)
            # chmod +x run_api.sh
            # ./run_api.sh
            ```
        -   **Windows (使用 Waitress):**
            ```batch
            # 示例: 监听 0.0.0.0:5000，使用4个线程
            waitress-serve --listen=0.0.0.0:5000 --threads=4 api_server:app
            # 或者使用提供的 run_api.bat (可能需要调整)
            # run_api.bat
            ```

3.  确保API服务正在运行并且ZeroBot插件可以访问到它 (例如，检查防火墙设置)。你可以访问 `http://<API_HOST>:<API_PORT>/health` 来检查API健康状态。

### 步骤 2: 编译和安装 Go ZeroBot 插件

**先决条件:**
-   Go 1.18 或更高版本。
-   已正确安装和配置的ZeroBot环境。

**安装:**

1.  将 `go_plugin` 目录复制到你的ZeroBot插件目录下 (通常是 `plugins/` 目录，具体路径取决于你的ZeroBot设置)。或者，如果你的ZeroBot支持通过 `go get` 或模块依赖方式加载插件，请遵循其规范。

2.  进入 `go_plugin` 目录:
    ```bash
    cd path/to/your/zerobot/plugins/go_plugin 
    # 或者你的项目根目录下的 go_plugin
    ```

3.  配置 `config.json`:
    创建一个 `config.json` 文件 (可以从 `go_plugin/config.json` 示例复制)，并根据你的设置进行修改：
    -   `jm_api_base_url`: **必须**指向你部署的Python API服务的地址 (例如, `"http://localhost:5000"` 或 `"http://your_api_server_ip:5000"`)。
    -   `jm_api_client_type`: API服务内部使用的JM客户端类型，通常为 `"html"`。
    -   `command_prefix`: 插件的命令前缀 (例如, `"jm"`)。
    -   `request_timeout_seconds`: Go插件调用API的超时时间。

4.  编译插件:
    如果ZeroBot需要编译好的插件文件 (如 `.so` 或可执行文件)，你需要根据ZeroBot的插件规范进行编译。
    通常，如果插件是一个独立的Go模块被ZeroBot加载，你可能只需要确保代码在正确的位置，ZeroBot会在启动时编译或加载它。
    如果需要手动编译 (例如，编译为 `.so` 文件给 `cza` 类型的插件):
    ```bash
    # 示例编译命令，具体请参考ZeroBot插件类型文档
    # go build -buildmode=plugin -o ../jmcomic.so . 
    ```
    对于很多现代Go框架，将插件作为主程序的一部分或通过Go模块依赖进行管理更为常见。请参考ZeroBot的文档。

5.  (重新)启动 ZeroBot。插件应该会被加载。

## 使用方法

假设你的命令前缀配置为 `"jm"`:

-   **帮助**: `jm help`
    显示插件的帮助信息和可用命令。

-   **搜索漫画**: `jm search <关键词>`
    例如: `jm search 老师`
    机器人会返回搜索结果列表，包含漫画标题和ID。

-   **查看详情**: `jm detail <漫画ID>`
    例如: `jm detail 12345` (这里的漫画ID从搜索结果中获取)
    机器人会返回漫画的详细信息，包括作者、标签、简介和章节列表 (包含章节ID)。

-   **下载章节**: `jm download <漫画ID> <章节ID1> [章节ID2...]`
    例如: `jm download 12345 67890 67891` (章节ID从详情中获取)
    机器人会向Python API服务提交下载请求。
    **注意**:
    -   下载操作在Python API服务器端进行。
    -   文件会保存在API服务器上 `jm.yaml` 中 `option.dir.base_dir` 配置的目录下。
    -   插件不会直接将文件发送给用户，用户需要有其他方式从API服务器获取下载的文件。

## 故障排除

-   **API服务无法启动**:
    -   检查Python和pip是否正确安装。
    -   确保虚拟环境已激活。
    -   查看 `requirements.txt` 中的依赖是否都已成功安装。
    -   检查 `jm.yaml` 配置是否正确，特别是域名和路径。
    -   查看API服务器启动时的控制台日志以获取错误信息。
-   **Go插件无法连接到API服务**:
    -   确保API服务正在运行。
    -   检查 `go_plugin/config.json` 中的 `jm_api_base_url` 是否正确指向API服务的地址和端口。
    -   检查网络连接和防火墙设置，确保ZeroBot运行的机器可以访问API服务。
-   **搜索/详情/下载失败**:
    -   查看ZeroBot的日志和Python API服务的日志，通常会有更详细的错误信息。
    -   可能是 `jm.yaml` 配置问题 (如域名失效、代理问题)。
    -   可能是JMComic网站本身的问题或反爬虫策略变更。

## 贡献

欢迎提交Issue和Pull Request。

## 许可证

MIT License (或者你选择的其他许可证)
