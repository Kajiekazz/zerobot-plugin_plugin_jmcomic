package jmcomic

import (
	"github.com/FloatTech/zerobot/common/plugin"
	zlog "github.com/FloatTech/zerobot/common/log"
	zero "github.com/FloatTech/zerobot/core"
)

// JMComicPlugin 结构体，可以包含插件的状态或配置（如果需要）
// 对于这个插件，大部分配置通过包级变量 cfg 管理
type JMComicPlugin struct{}

// Name 返回插件名
func (p *JMComicPlugin) Name() string {
	return pluginName
}

// Author 返回作者名
func (p *JMComicPlugin) Author() string {
	return pluginAuthor
}

// Version 返回版本号
func (p *JMComicPlugin) Version() string {
	return pluginVersion
}

// Description 返回描述
func (p *JMComicPlugin) Description() string {
	return pluginDesc
}

// OnLoad 插件加载时执行的函数
// 在这里进行命令注册等初始化操作
func (p *JMComicPlugin) OnLoad(e *zero.Engine) {
	zlog.Infof("[%s] OnLoad called. Registering handlers...", pluginName)
	// 确保httpClient的超时已经根据cfg设置
	if httpClient.Timeout != cfg.timeoutDuration {
		httpClient.Timeout = cfg.timeoutDuration
		zlog.Debugf("[%s] HTTP client timeout set to %v in OnLoad", pluginName, cfg.timeoutDuration)
	}
	MustRegisterHandlers(e) // 将引擎实例传递给处理器注册函数
	zlog.Infof("[%s] Plugin (v%s by %s) loaded and handlers registered.", pluginName, pluginVersion, pluginAuthor)
}

// OnUnload 插件卸载时执行的函数 (可选)
func (p *JMComicPlugin) OnUnload(e *zero.Engine) {
	zlog.Infof("[%s] Plugin unloaded.", pluginName)
}

// init 函数在包被导入时执行，用于注册插件
func init() {
	// 创建插件实例并注册
	// ZeroBot v0.8.x 推荐使用 plugin.Register
	err := plugin.Register(&JMComicPlugin{})
	if err != nil {
		zlog.Fatalf("[%s] 插件注册失败: %v", pluginName, err)
	}
	zlog.Infof("[%s] 插件已在 init 中提交注册。", pluginName)
}
