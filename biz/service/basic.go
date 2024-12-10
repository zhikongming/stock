package service

import (
	"context"
	"io"
	"log"
	"net/http"
)

func DoGet(ctx context.Context, url string, params map[string]string, headers map[string]string) ([]byte, error) {
	client := &http.Client{}

	// 创建一个新的 http.Request 对象
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("NewRequest error: %v", err)
		return nil, err
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 设置参数
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Do request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取并打印响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read bytes: %v", err)
		return nil, err
	}

	return body, nil
}
