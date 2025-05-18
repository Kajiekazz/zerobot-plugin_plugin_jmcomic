package jmcomic

import (
    "encoding/json"
	"fmt"
	"net/http"
	"time"
)

// 执行JM查询
func queryJM(id string) (*jmResponse, error) {
	var (
		lastErr error
		result  *jmResponse
	)

	// 域名故障转移重试
	for _, domain := range currentDom {
		for i := 0; i < cfg.RetryTimes; i++ {
			if result, lastErr = tryQuery(domain, id); lastErr == nil {
				return result, nil
			}
			time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
		}
	}

	// 域名自动切换
	if rotateDomain() {
		return queryJM(id) // 递归重试
	}

	return nil, fmt.Errorf("所有域名不可用 | 最后错误：%v", lastErr)
}

// 尝试查询单个域名
func tryQuery(domain, id string) (*jmResponse, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/api/v1/album/%s", domain, id), nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result jmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API返回错误码：%d", result.Code)
	}
	return &result, nil
}

// 轮换域名
func rotateDomain() bool {
	domainLock.Lock()
	defer domainLock.Unlock()

	if len(currentDom) == 0 {
		return false
	}
	
	// 移除失效域名
	currentDom = currentDom[1:]
	domainIdx = 0
	
	return len(currentDom) > 0
}