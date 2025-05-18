package jmcomic

import (
	"context"
	"fmt"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	// 注册测试命令
	engine := control.Register("jmtest", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "域名测试",
		Help:             "发送【测试禁漫域名】",
	})

	engine.OnFullMatch("测试禁漫域名").Handle(func(ctx *zero.Ctx) {
		ctx.Send("开始域名健康检查...")
		
		var (
			wg      sync.WaitGroup
			results = make(chan string, len(cfg.Domains))
		)
		// 并发测试所有域名
		for _, domain := range cfg.Domains {
			wg.Add(1)
			go func(d string) {
				defer wg.Done()
				testDomain(d, results)
			}(domain)
		}

		// 收集结果
		go func() {
			wg.Wait()
			close(results)
			
			report := "【域名测试报告】\n"
			for res := range results {
				report += res + "\n"
			}
			ctx.Send(report)
		}()
	})
}

// 测试单个域名
func testDomain(domain string, results chan<- string) {
	start := time.Now()
	_, err := tryQuery(domain, "test123") // 测试用ID

	status := "✅ 正常"
	if err != nil {
		status = fmt.Sprintf("❌ 失败 (%v)", err)
	}

	results <- fmt.Sprintf("%-20s %-15s 耗时：%v", 
		domain, 
		status, 
		time.Since(start).Round(time.Millisecond),
	)
}