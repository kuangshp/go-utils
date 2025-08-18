package k

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
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

func GetIpToAddress(ip string) (province, city string, origin IPInfoDataDetail) {
	httpUrl := fmt.Sprintf("https://ip.taobao.com/getIpInfo.php?ip=%s", ip)
	response, err := http.Get(httpUrl)
	if err != nil {
		log.Fatalf("发起请求失败:%v", err)
		return "", "", IPInfoDataDetail{}
	}
	defer func() {
		if err1 := response.Body.Close(); err1 != nil {
			return
		}
	}()
	readAll, err1 := io.ReadAll(response.Body)
	if err1 != nil {
		log.Fatalf("读取请求返回失败:%v", err1)
	}
	fmt.Println(string(readAll))
	var ipInfo IPInfoResponse
	if err := json.Unmarshal(readAll, &ipInfo); err != nil {
		log.Printf("解析JSON失败: %v", err)
		return "", "", IPInfoDataDetail{}
	}
	// 检查返回码
	if ipInfo.Code == "0" {
		return ipInfo.Data.PROVINCE_CN, ipInfo.Data.CITY_CN, ipInfo.Data
	}
	return "", "", IPInfoDataDetail{}
}
