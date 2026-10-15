package k

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

type IPInfoDataDetail struct {
	CITY_EN       string `json:"CITY_EN"`
	QUERY_IP      string `json:"QUERY_IP"`
	CITY_CODE     string `json:"CITY_CODE"`
	CITY_CN       string `json:"CITY_CN"`
	COUNTY_EN     string `json:"COUNTY_EN"`
	LONGITUDE     string `json:"LONGITUDE"`
	PROVINCE_CN   string `json:"PROVINCE_CN"`
	TZONE         string `json:"TZONE"`
	PROVINCE_EN   string `json:"PROVINCE_EN"`
	ISP_EN        string `json:"ISP_EN"`
	AREA_CODE     string `json:"AREA_CODE"`
	PROVINCE_CODE string `json:"PROVINCE_CODE"`
	ISP_CN        string `json:"ISP_CN"`
	AREA_CN       string `json:"AREA_CN"`
	COUNTRY_CN    string `json:"COUNTRY_CN"`
	AREA_EN       string `json:"AREA_EN"`
	COUNTRY_EN    string `json:"COUNTRY_EN"`
	COUNTY_CN     string `json:"COUNTY_CN"`
	COUNTY_CODE   string `json:"COUNTY_CODE"`
	ASN           string `json:"ASN"`
	LATITUDE      string `json:"LATITUDE"`
	COUNTRY_CODE  string `json:"COUNTRY_CODE"`
	ISP_CODE      string `json:"ISP_CODE"`
}
type IPInfoResponse struct {
	Code string           `json:"code"`
	Data IPInfoDataDetail `json:"data"`
}

const defaultIPInfoURLFmt = "https://ip.taobao.com/getIpInfo.php?ip=%s"

var defaultIPHTTPClient = &http.Client{Timeout: 5 * time.Second}

// LookupIpAddress 查询 IP 归属地，失败返回 error，不会终止进程。
func LookupIpAddress(ip string) (province, city string, origin IPInfoDataDetail, err error) {
	return LookupIpAddressURL(defaultIPHTTPClient, fmt.Sprintf(defaultIPInfoURLFmt, ip))
}

// LookupIpAddressURL 使用指定 HTTP 客户端与完整 URL 查询归属地，便于测试注入 httptest。
func LookupIpAddressURL(client *http.Client, requestURL string) (province, city string, origin IPInfoDataDetail, err error) {
	if client == nil {
		client = defaultIPHTTPClient
	}
	response, err := client.Get(requestURL)
	if err != nil {
		return "", "", IPInfoDataDetail{}, fmt.Errorf("发起 IP 归属查询失败: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", IPInfoDataDetail{}, fmt.Errorf("读取 IP 归属响应失败: %w", err)
	}

	var ipInfo IPInfoResponse
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return "", "", IPInfoDataDetail{}, fmt.Errorf("解析 IP 归属 JSON 失败: %w", err)
	}
	if ipInfo.Code != "0" {
		return "", "", IPInfoDataDetail{}, fmt.Errorf("IP 归属查询返回码异常: %s", ipInfo.Code)
	}
	return ipInfo.Data.PROVINCE_CN, ipInfo.Data.CITY_CN, ipInfo.Data, nil
}

// GetIpToAddress 查询 IP 归属地。
//
// Deprecated: 请改用 LookupIpAddress。本函数保留旧签名且永不 Fatal/Exit；失败时返回空值。
func GetIpToAddress(ip string) (province, city string, origin IPInfoDataDetail) {
	province, city, origin, err := LookupIpAddress(ip)
	if err != nil {
		return "", "", IPInfoDataDetail{}
	}
	return province, city, origin
}
