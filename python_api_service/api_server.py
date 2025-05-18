import os
import sys
import logging
from flask import Flask, request, jsonify, abort
from jmcomic import create_option, JmHtmlClient, JmApiClient, JmImageClient, JmDownloader, JmcomicText

# 将当前脚本所在目录添加到sys.path，以便jmcomic能正确找到配置文件等
# 如果jm.yaml与api_server.py在同一目录，通常jmcomic可以自动找到
# 但为了保险起见，可以这样做：
# current_dir = os.path.dirname(os.path.abspath(__file__))
# if current_dir not in sys.path:
#    sys.path.append(current_dir)


# 配置日志
log_format = '%(asctime)s - %(levelname)s - %(message)s'
logging.basicConfig(level=logging.INFO, format=log_format)
# 减少jmcomic库的冗余日志，只显示错误以上级别
for logger_name in ['jmcomic', 'jmcomic.jm_client', 'jmcomic.jm_downloader', 'jmcomic.jm_option', 'jmcomic.jm_entity']:
    logging.getLogger(logger_name).setLevel(logging.ERROR)

app = Flask(__name__)

# 全局Option对象，在应用启动时加载一次
# jmcomic库会从当前工作目录或用户目录查找 jm.yaml
# 确保 api_server.py 运行时，其工作目录是 python_api_service/
try:
    # 明确指定配置文件路径为脚本同目录下的 jm.yaml
    config_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'jm.yaml')
    if not os.path.exists(config_path):
        logging.warning(f"jm.yaml not found at {config_path}. Using default jmcomic settings.")
        # 如果没有jm.yaml，jmcomic会使用内置的默认配置，可能无法正常工作
        # 或者你可以选择在这里强制退出或使用一个保底的配置
        # sys.exit("Error: jm.yaml not found. Please configure it.")
        GLOBAL_OPTION = create_option(None) # 尝试默认加载
    else:
        GLOBAL_OPTION = create_option(config_path)
    logging.info(f"Loaded jm.yaml from: {GLOBAL_OPTION.get_config_file_path()}")
    # 确保下载目录存在
    os.makedirs(GLOBAL_OPTION.dir_rule.base_dir, exist_ok=True)
    logging.info(f"Comic download base directory: {GLOBAL_OPTION.dir_rule.base_dir}")

except Exception as e:
    logging.error(f"Failed to initialize JMComic option: {e}", exc_info=True)
    # 如果初始化失败，后续的API调用很可能会出错
    GLOBAL_OPTION = None # 或者抛出异常使应用启动失败

def get_client(client_type='html'):
    if GLOBAL_OPTION is None:
        raise ConnectionRefusedError("JMComic global option not initialized. Check jm.yaml.")
    # 每次请求都创建一个新的client实例，使用全局option
    # 这样可以避免多线程环境下client状态共享的问题（虽然Flask默认单线程处理请求，但生产环境gunicorn/waitress会多worker）
    option_copy = GLOBAL_OPTION # jmcomic的option设计上应该是可以复用的
    if client_type == 'api':
        return JmApiClient(option_copy)
    return JmHtmlClient(option_copy)

def get_image_client():
    if GLOBAL_OPTION is None:
        raise ConnectionRefusedError("JMComic global option not initialized. Check jm.yaml.")
    option_copy = GLOBAL_OPTION
    return JmImageClient(option_copy)


@app.route('/health', methods=['GET'])
def health_check():
    if GLOBAL_OPTION is None:
        return jsonify({"status": "error", "message": "JMComic option not initialized"}), 500
    return jsonify({"status": "ok", "message": "API service is running"}), 200

@app.route('/search', methods=['GET'])
def search_comic_api():
    keywords = request.args.get('keyword')
    client_type = request.args.get('client_type', 'html') # 默认html
    if not keywords:
        return jsonify({"status": "error", "message": "Missing 'keyword' parameter"}), 400

    try:
        client = get_client(client_type)
        logging.info(f"Searching for: '{keywords}' using {client_type} client")
        results = list(client.search_album(keywords)) # search_album 返回生成器
        output = []
        for album_image in results: # AlbumImage 对象
            output.append({
                'id': album_image.id,
                'title': JmcomicText.parse_text(album_image.title), # 确保文本可读
                'author': ", ".join(JmcomicText.parse_list(album_image.author_list)),
                'tags': ", ".join(JmcomicText.parse_list(album_image.tag_list)),
                'description': JmcomicText.parse_text(album_image.description) if hasattr(album_image, 'description') else "N/A",
                'cover_url': album_image.cover_url if hasattr(album_image, 'cover_url') else None,
                'source_site': album_image.source_site if hasattr(album_image, 'source_site') else "N/A"
            })
        logging.info(f"Search for '{keywords}' found {len(output)} results.")
        return jsonify({"status": "success", "data": output})
    except Exception as e:
        logging.error(f"Error during search for '{keywords}': {e}", exc_info=True)
        return jsonify({"status": "error", "message": str(e)}), 500

@app.route('/comic/<album_id>', methods=['GET'])
def get_comic_detail_api(album_id):
    client_type = request.args.get('client_type', 'html')
    if not album_id:
        return jsonify({"status": "error", "message": "Missing 'album_id' in path"}), 400

    try:
        client = get_client(client_type)
        logging.info(f"Fetching details for album_id: '{album_id}' using {client_type} client")
        album_detail = client.get_album_detail(album_id) # AlbumDetail 对象

        chapters_output = []
        if hasattr(album_detail, 'chapter_list'):
            for chapter_image in album_detail.chapter_list: # ChapterImage 对象
                chapters_output.append({
                    'id': chapter_image.id,
                    'title': JmcomicText.parse_text(chapter_image.title),
                    'index': str(chapter_image.index) if hasattr(chapter_image, 'index') else "N/A",
                    'page_count': chapter_image.page_count if hasattr(chapter_image, 'page_count') else 0,
                })
        
        detail_data = {
            'id': album_detail.id,
            'title': JmcomicText.parse_text(album_detail.title),
            'author': ", ".join(JmcomicText.parse_list(album_detail.author_list)),
            'tags': ", ".join(JmcomicText.parse_list(album_detail.tag_list)),
            'description': JmcomicText.parse_text(album_detail.description),
            'cover_url': album_detail.cover_url,
            'chapters': chapters_output,
            'source_site': album_detail.source_site if hasattr(album_detail, 'source_site') else "N/A"
        }
        logging.info(f"Successfully fetched details for album_id: '{album_id}'.")
        return jsonify({"status": "success", "data": detail_data})
    except Exception as e:
        logging.error(f"Error fetching detail for album_id '{album_id}': {e}", exc_info=True)
        return jsonify({"status": "error", "message": str(e)}), 500

@app.route('/download/<album_id>', methods=['POST'])
def download_chapters_api(album_id):
    if not request.is_json:
        return jsonify({"status": "error", "message": "Request body must be JSON"}), 400
    
    data = request.get_json()
    chapter_ids = data.get('chapter_ids') # 期望是一个字符串列表

    if not chapter_ids or not isinstance(chapter_ids, list):
        return jsonify({"status": "error", "message": "Missing or invalid 'chapter_ids' (must be a list of strings)"}), 400
    
    try:
        # 下载通常使用 JmImageClient
        image_client = get_image_client()
        logging.info(f"Requesting download for album_id: '{album_id}', chapters: {chapter_ids}")
        
        # JmImageClient.download_album可以直接处理下载
        # 它会将文件下载到 GLOBAL_OPTION.dir_rule.base_dir 下的结构化目录中
        image_client.download_album(album_id, include_chapters=chapter_ids)
        
        # 获取漫画对象以构建下载路径提示 (可选)
        # album_obj_for_path = image_client.get_album_detail(album_id) # 这会再次请求详情
        # download_path_hint = GLOBAL_OPTION.dir_rule.comic_image_dir(
        #     GLOBAL_OPTION.dir_rule.base_dir,
        #     album_obj_for_path.title # 假设有title属性
        # )
        # 简化：提示用户检查API服务器上的配置下载目录
        download_path_hint = f"Check API server's configured download directory: {GLOBAL_OPTION.dir_rule.base_dir}"


        logging.info(f"Download task for album '{album_id}', chapters {chapter_ids} submitted.")
        return jsonify({
            "status": "success",
            "message": f"漫画 {album_id} 的章节 {chapter_ids} 已提交下载请求。",
            "download_path_hint": download_path_hint
        })
    except Exception as e:
        logging.error(f"Error submitting download for album_id '{album_id}', chapters {chapter_ids}: {e}", exc_info=True)
        return jsonify({"status": "error", "message": str(e)}), 500


if __name__ == '__main__':
    # 从环境变量获取端口和主机，方便Docker等部署
    host = os.environ.get('API_HOST', '0.0.0.0')
    port = int(os.environ.get('API_PORT', 5000)) # Flask默认端口5000
    # debug=True 只用于开发环境，生产环境应使用Gunicorn或Waitress
    # app.run(host=host, port=port, debug=True)
    # 生产环境建议:
    # Linux/macOS: gunicorn -w 4 -b 0.0.0.0:5000 api_server:app
    # Windows: waitress-serve --listen=0.0.0.0:5000 api_server:app
    # 这里为了简单，直接运行，但提示用户使用生产级服务器
    logging.info(f"Starting Flask API server on {host}:{port}")
    logging.info("For production, use Gunicorn (Linux/macOS) or Waitress (Windows).")
    app.run(host=host, port=port, debug=False) # debug=False for default run without Gunicorn

