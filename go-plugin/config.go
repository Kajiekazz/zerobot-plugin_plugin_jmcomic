package jmcomic

import (
	"time"

	"github.com/FloatTech/zerobot/common/config"
	zlog "github.com/FloatTech/zerobot/common/log"
)

const (
	pluginName   = "jmcomic" // 插件名称，用于配置和日志
	pluginVersion = "1.2.0"   // 插件版本
	pluginAuthor  = "YourName" // 替换为你的名字
	pluginDesc    = "通过API与JMComic交互，提供搜索、详情和下载请求功能。"
)

// PluginConfig 定义插件配置结构
type PluginConfig struct {
	ApiBaseURL              string `json:"api_base_url"`
	ApiClientType           string `json:"api_client_type"`
	RequestTimeoutSeconds   int    `json:"request_timeout_seconds"`
	MaxSearchResultsDisplay int    `json:"max_search_results_display"`
	MaxChaptersDisplay      int    `json:"max_chapters_display"`
	// CommandPrefix string `json:"command_prefix"` // 如果不再需要可配置前缀，可以移除

	// 内部使用
	timeoutDuration time.Duration
}

var cfg = &PluginConfig{ // 默认配置
	ApiBaseURL:              "http://localhost:5000",
	ApiClientType:           "html",
	RequestTimeoutSeconds:   30,
	MaxSearchResultsDisplay: 5,
	MaxChaptersDisplay:      10,
	// CommandPrefix:           "jm",
}

func init() {
	// 从ZeroBot的配置系统中加载配置
	err := config.GetConfig(pluginName, cfg)
	if err != nil {
		zlog.Errorf("[%s] 加载配置文件失败: %v, 将使用默认配置", pluginName, err)
	} else {
		zlog.Infof("[%s] 配置文件已加载", pluginName)
	}

	// 根据加载的配置更新内部使用的值
	if cfg.RequestTimeoutSeconds <= 0 {
		cfg.RequestTimeoutSeconds = 30
	}
	cfg.timeoutDuration = time.Duration(cfg.RequestTimeoutSeconds) * time.Second

	if cfg.MaxSearchResultsDisplay <= 0 {
		cfg.MaxSearchResultsDisplay = 5
	}
	if cfg.MaxChaptersDisplay <= 0 {
		cfg.MaxChaptersDisplay = 10
	}
	// if cfg.CommandPrefix == "" { // 如果仍然保留CommandPrefix字段
	// 	cfg.CommandPrefix = "jm"
	// }
}
