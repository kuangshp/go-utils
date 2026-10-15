package k

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	if got := ClientIP(req); got != "1.2.3.4" {
		t.Fatalf("X-Forwarded-For = %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-Ip", "9.9.9.9")
	if got := ClientIP(req); got != "9.9.9.9" {
		t.Fatalf("X-Real-Ip = %q", got)
	}
}

func TestLookupIpAddressURL_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"0","data":{"PROVINCE_CN":"广东","CITY_CN":"深圳","QUERY_IP":"1.1.1.1"}}`))
	}))
	defer server.Close()

	province, city, origin, err := LookupIpAddressURL(server.Client(), server.URL)
	if err != nil {
		t.Fatalf("LookupIpAddressURL: %v", err)
	}
	if province != "广东" || city != "深圳" || origin.QUERY_IP != "1.1.1.1" {
		t.Fatalf("got province=%s city=%s origin=%+v", province, city, origin)
	}
}

func TestLookupIpAddressURL_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer server.Close()

	// 非 JSON 体会导致解析失败，进程必须存活
	_, _, _, err := LookupIpAddressURL(server.Client(), server.URL)
	if err == nil {
		t.Fatal("期望错误")
	}
}

func TestLookupIpAddressURL_BadCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":"1","data":{}}`))
	}))
	defer server.Close()

	_, _, _, err := LookupIpAddressURL(server.Client(), server.URL)
	if err == nil {
		t.Fatal("异常返回码应报错")
	}
}
