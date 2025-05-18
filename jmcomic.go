package jmcomic

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 配置结构体
type config struct {
	Proxy      string   `yaml:"proxy"`      // 代理地址格式：http://ip:port
	RetryTimes int      `yaml:"retryTimes"` // 单域名重试次数
	Domains    []string `yaml:"domains"`    // 域名列表
	CacheSize  int      `yaml:"cacheSize"`  // 缓存容量
}

// 默认配置
var cfg = config{
	Proxy:      "http://127.0.0.1:7890",
	RetryTimes: 3,
	Domains: []string{
		"jmcomic1.me",
		"18comic.vip",
		"jmcomic.me",
	},
	CacheSize: 100,
}

// API响应结构
type jmResponse struct {
	Code int `json:"code"`
	Data struct {
		Title  string   `json:"title"`
		Author string   `json:"author"`
		Tags   []string `json:"tags"`
	} `json:"data"`
}

// 缓存条目
type cacheItem struct {
	Data      *jmResponse
	ExpiresAt time.Time
}

var (
	client     *http.Client      // HTTP客户端
	cache      sync.Map          // 线程安全缓存
	domainLock sync.Mutex        // 域名锁
	currentDom []string          // 当前可用域名
	domainIdx  int               // 当前域名索引
)

func init() {
	// 初始化配置
	initConfig()
	
	// 创建HTTP客户端
	initHTTPClient()
	
	// 注册插件
	engine := control.Register("jmcomic", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "禁漫ID查询",
		Help:             "发送 JM+数字ID 查询本子信息\n示例：JM123456",
	})

	// 注册事件处理器
	engine.OnRegex(`^(?i)jm\d+$`).SetBlock(true).Handle(handleQuery)
}

// 初始化配置
func initConfig() {
	// 环境变量覆盖
	if p := os.Getenv("JM_PROXY"); p != "" {
		cfg.Proxy = p
	}
	currentDom = cfg.Domains
}

// 初始化HTTP客户端
func initHTTPClient() {
	proxyURL, _ := url.Parse(cfg.Proxy)
	client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 15 * time.Second,
	}
}

// 处理查询请求
func handleQuery(ctx *zero.Ctx) {
	// 提取ID
	id := ctx.State["regex_matched"].([]string)[0][2:]
	
	// 检查缓存
	if v, ok := cache.Load(id); ok {
		if item := v.(cacheItem); time.Now().Before(item.ExpiresAt) {
			sendResult(ctx, item.Data)
			return
		}
	}

	// 执行查询
	result, err := queryJM(id)
	if err != nil {
		ctx.SendChain(message.Text("查询失败：", err))
		return
	}

	// 缓存结果
	cache.Store(id, cacheItem{
		Data:      result,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	// 发送结果
	sendResult(ctx, result)
}

// 发送结果消息
func sendResult(ctx *zero.Ctx, data *jmResponse) {
    msg := []message.Segment{
        message.Text(fmt.Sprintf("标题：%s\n", filterText(data.Data.Title))), // 修复括号闭合
        message.Text(fmt.Sprintf("作者：%s\n", data.Data.Author)),          // 确保每行有逗号
        message.Text(fmt.Sprintf("标签：%v\n", data.Data.Tags)),            // 同上
        message.Text("⚠ 部分内容可能被自动过滤"),                            // 最后一个元素可省略逗号
    }
    ctx.Send(msg) 
}
// 敏感词过滤
func filterText(text string) string {
	// 示例过滤逻辑（需自行完善）
	sensitive := []string{"敏感词1", "敏感词2"}
	for _, word := range sensitive {
		text = regexp.MustCompile(word).ReplaceAllString(text, "***")
	}
	return text
}