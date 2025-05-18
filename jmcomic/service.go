package jmcomic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	// "time" // 如果 init 中不再设置 httpClient.Timeout，则可能不需要

	zlog "github.com/FloatTech/zerobot/common/log"
)

var httpClient = &http.Client{} // 全局HTTP客户端

func init() {
	// httpClient.Timeout 的设置移到 config.go 的 init 中，因为 cfg 在那里最终确定
	// 或者，如果 Plugin 结构体有 OnLoad 方法，可以在 OnLoad 中设置
	// 这里假设 cfg.timeoutDuration 在 service.go 的函数调用时已经正确设置
}

// makeAPIRequest 发起HTTP请求到Python API服务
func makeAPIRequest(ctx context.Context, method, endpoint string, queryParams map[string]string, body interface{}) (*APIResponse, error) {
	if cfg.ApiBaseURL == "" {
		return nil, fmt.Errorf("API基础URL未在配置中设置")
	}
	// 确保httpClient的超时已设置 (可以在OnLoad中做，或者每次请求前检查并应用)
	// 为了简单，我们依赖config.go中init的设置
	if httpClient.Timeout != cfg.timeoutDuration { // 如果超时可能变化，则每次更新
		httpClient.Timeout = cfg.timeoutDuration
	}


	fullURL, err := url.Parse(cfg.ApiBaseURL)
	if err != nil {
		zlog.Errorf("[%s API Call] 解析基础URL '%s' 失败: %v", pluginName, cfg.ApiBaseURL, err)
		return nil, fmt.Errorf("无效的API基础URL: %w", err)
	}
	fullURL.Path = strings.TrimRight(fullURL.Path, "/") + "/" + strings.TrimLeft(endpoint, "/")

	q := fullURL.Query()
	if queryParams != nil {
		for k, v := range queryParams {
			q.Set(k, v)
		}
	}
	q.Set("client_type", cfg.ApiClientType)
	fullURL.RawQuery = q.Encode()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			zlog.Errorf("[%s API Call] 序列化请求体失败: %v", pluginName, err)
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		zlog.Errorf("[%s API Call] 创建HTTP请求失败: %v", pluginName, err)
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "ZeroBot-JMComic-Plugin/"+pluginVersion)

	zlog.Debugf("[%s API Call] Request: %s %s", pluginName, method, fullURL.String())

	resp, err := httpClient.Do(req)
	if err != nil {
		zlog.Errorf("[%s API Call] HTTP请求执行失败 (%s %s): %v", pluginName, method, fullURL.String(), err)
		return nil, fmt.Errorf("API请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		zlog.Errorf("[%s API Call] 读取响应体失败: %v", pluginName, err)
		return nil, fmt.Errorf("读取API响应体失败: %w", err)
	}
	zlog.Debugf("[%s API Call] Response Status: %s, Body: %s", pluginName, resp.Status, string(respBody))

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		zlog.Errorf("[%s API Call] 解析API响应JSON失败: %v. Raw body: %s", pluginName, err, string(respBody))
		return &APIResponse{Status: "error", Message: fmt.Sprintf("无法解析API响应 (HTTP %d): %s", resp.StatusCode, string(respBody))},
			fmt.Errorf("解析API响应JSON失败: %w (raw: %s)", err, string(respBody))
	}

	if resp.StatusCode >= 400 {
		zlog.Errorf("[%s API Call] API返回HTTP错误 %d: %s", pluginName, resp.StatusCode, apiResp.Message)
		return &apiResp, fmt.Errorf("API错误 (HTTP %d): %s", resp.StatusCode, apiResp.Message)
	}

	if apiResp.Status == "error" {
		zlog.Warnf("[%s API Call] API业务逻辑错误: %s", pluginName, apiResp.Message)
		return &apiResp, fmt.Errorf("API业务错误: %s", apiResp.Message)
	}

	return &apiResp, nil
}

// SearchComic 调用API搜索漫画
func SearchComic(ctx context.Context, keyword string) ([]ComicSearchResultItem, error) {
	params := map[string]string{"keyword": keyword}
	apiResp, err := makeAPIRequest(ctx, http.MethodGet, "/search", params, nil)
	if err != nil {
		return nil, err
	}

	var results []ComicSearchResultItem
	if err := json.Unmarshal(apiResp.Data, &results); err != nil {
		zlog.Errorf("[%s Service] 解析搜索结果数据失败: %v", pluginName, err)
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}
	return results, nil
}

// GetComicDetail 调用API获取漫画详情
func GetComicDetail(ctx context.Context, albumID string) (*ComicDetail, error) {
	endpoint := fmt.Sprintf("/comic/%s", albumID)
	apiResp, err := makeAPIRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var detail ComicDetail
	if err := json.Unmarshal(apiResp.Data, &detail); err != nil {
		zlog.Errorf("[%s Service] 解析详情数据失败: %v", pluginName, err)
		return nil, fmt.Errorf("解析漫画详情失败: %w", err)
	}
	return &detail, nil
}

// DownloadChapters 调用API下载章节
func DownloadChapters(ctx context.Context, albumID string, chapterIDs []string) (string, string, error) {
	endpoint := fmt.Sprintf("/download/%s", albumID)
	reqBody := DownloadRequest{ChapterIDs: chapterIDs}

	apiResp, err := makeAPIRequest(ctx, http.MethodPost, endpoint, nil, reqBody)
	if err != nil {
		errMsg := "下载请求失败"
		if apiResp != nil && apiResp.Message != "" {
			errMsg = apiResp.Message
		} else if err != nil {
			errMsg = err.Error()
		}
		return errMsg, "", err
	}
	return apiResp.Message, apiResp.DownloadPathHint, nil
}
