package k

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	client, err := NewHTTPClient("https://ip.taobao.com/getIpInfo.php")
	if err != nil {
		fmt.Println("创建HTTP客户端失败:", err)
	}
	response, err := client.Get("", map[string]string{
		"ip": "",
	})
	all, err := client.ReadBodyString(response)
	fmt.Println(string(all))
}

func TestPostJSON(t *testing.T) {
	client, err := NewHTTPClient("http://xx.com/api/v1/admin/")
	if err != nil {
		fmt.Println("创建HTTP客户端失败:", err)
	}
	response, err := client.PostJSON("auth/login", map[string]interface{}{
		"username": "admin",
		"password": "123456",
	})
	all, err := client.ReadBodyString(response)
	fmt.Println(string(all))
}
