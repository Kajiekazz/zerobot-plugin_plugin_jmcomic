package jmcomic

import "encoding/json"

// APIResponse 通用API响应结构
type APIResponse struct {
	Status           string          `json:"status"` // "success" or "error"
	Data             json.RawMessage `json:"data,omitempty"`
	Message          string          `json:"message,omitempty"`
	DownloadPathHint string          `json:"download_path_hint,omitempty"`
}

// ComicSearchResultItem 搜索结果中的单个漫画项
type ComicSearchResultItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Tags        string `json:"tags"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	SourceSite  string `json:"source_site"`
}

// ChapterInfo 漫画章节信息
type ChapterInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Index     string `json:"index"`
	PageCount int    `json:"page_count"`
}

// ComicDetail 漫画详细信息
type ComicDetail struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Author      string        `json:"author"`
	Tags        string        `json:"tags"`
	Description string        `json:"description"`
	CoverURL    string        `json:"cover_url"`
	Chapters    []ChapterInfo `json:"chapters"`
	SourceSite  string        `json:"source_site"`
}

// DownloadRequest 下载请求体
type DownloadRequest struct {
	ChapterIDs []string `json:"chapter_ids"`
}
