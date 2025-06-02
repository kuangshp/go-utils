package k

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HTTPClient 是一个自定义的 HTTP 请求客户端
type HTTPClient struct {
	HttpUrl string
	Client  *http.Client
}

// NewHTTPClient 创建一个新的 HTTPClient 实例
// httpUrl 是目标服务器的地址，例如 "http://localhost:8080/api/v1"
// options 是可选的配置参数，例如代理 IP 超时时间
func NewHTTPClient(httpUrl string, options ...string) (*HTTPClient, error) {
	// 校验 httpUrl 是否是有效的 URL
	_, err := url.ParseRequestURI(httpUrl)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	if len(options) > 0 {
		switch len(options) {
		case 1: // 只设置了代理 IP
			// 如果提供了代理 IP，则设置代理
			proxyIp := options[0]
			if proxyIp == "" {
				return &HTTPClient{
					HttpUrl: httpUrl,
					Client:  client,
				}, nil // 如果没有提供代理 IP，则返回默认的 HTTPClient
			}
			proxy, err := url.Parse("http://" + strings.TrimSpace(string(proxyIp)))
			if err != nil {
				return &HTTPClient{
					HttpUrl: httpUrl,
					Client:  client,
				}, nil
			}
			client = &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
				Timeout:       6 * time.Second,
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxy),
				}, // 使用代理
			}
		case 2: // 设置了代理 IP 和超时时间
			proxyIp := options[0]
			if proxyIp == "" {
				return &HTTPClient{
					HttpUrl: httpUrl,
					Client:  client,
				}, nil // 如果没有提供代理 IP，则返回默认的 HTTPClient
			}
			proxy, err := url.Parse("http://" + strings.TrimSpace(string(proxyIp)))
			if err != nil {
				return &HTTPClient{
					HttpUrl: httpUrl,
					Client:  client,
				}, nil
			}
			timeout, err := strconv.Atoi(options[1])
			if err != nil {
				timeout = 6 // 如果转换失败，使用默认的 6 秒超时
			}
			client = &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
				Timeout:       time.Duration(timeout) * time.Second,
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxy),
				}, // 使用代理
			}
		}
	}
	return &HTTPClient{
		HttpUrl: httpUrl,
		Client:  client,
	}, nil
}

// buildURL 构建完整的请求 URL
func (c *HTTPClient) buildURL(path string) string {
	if path == "" {
		return c.HttpUrl
	}
	// 确保 path 以 "/" 开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.HttpUrl + path
}

// Get 发送一个 GET 请求
// path 是相对于 BasePath 的路径
// queryParams 是 URL 查询参数
func (c *HTTPClient) Get(path string, queryParams map[string]string) (*http.Response, error) {
	fullURL := c.buildURL(path)
	fmt.Println("Request URL:", fullURL)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	return c.Client.Do(req)
}

// Post 发送一个 POST 请求
// path 是相对于 BasePath 的路径
// body 是请求体，通常是 io.Reader 类型，例如 strings.NewReader(jsonData)
// contentType 是请求体的类型，例如 "application/json"
func (c *HTTPClient) Post(path string, body io.Reader, contentType string) (*http.Response, error) {
	fullURL := c.buildURL(path)

	req, err := http.NewRequest("POST", fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return c.Client.Do(req)
}

// Put 发送一个 PUT 请求
// path 是相对于 BasePath 的路径
// body 是请求体
// contentType 是请求体的类型
func (c *HTTPClient) Put(path string, body io.Reader, contentType string) (*http.Response, error) {
	fullURL := c.buildURL(path)

	req, err := http.NewRequest("PUT", fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return c.Client.Do(req)
}

// Delete 发送一个 DELETE 请求
// path 是相对于 BasePath 的路径
func (c *HTTPClient) Delete(path string) (*http.Response, error) {
	fullURL := c.buildURL(path)

	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return nil, err
	}

	return c.Client.Do(req)
}

// PostJSON 发送一个POST请求，使用map作为JSON请求体
// path 是相对于BasePath的路径
// data 是将被转换为JSON的map
func (c *HTTPClient) PostJSON(path string, data map[string]interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.Post(path, strings.NewReader(string(jsonData)), "application/json")
}

// ReadBody 读取响应体并返回字节数组
func (c *HTTPClient) ReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ReadBodyString 读取响应体并返回字符串
func (c *HTTPClient) ReadBodyString(resp *http.Response) (string, error) {
	body, err := c.ReadBody(resp)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ReadJSON 将响应体解析为指定的结构体
// data 应该是一个指向结构体的指针，例如 &User{}
func (c *HTTPClient) ReadJSON(resp *http.Response, data interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(data)
}
