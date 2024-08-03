package k

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestHttpGet(t *testing.T) {
	url := "https://jsonplaceholder.typicode.com/todos/1"
	header := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := HttpGet(url, header)
	if err != nil {
		t.Log("net error: " + err.Error())
		return
	}

	body, _ := io.ReadAll(resp.Body)
	t.Log("response: ", resp.StatusCode, string(body))
}

func TestHttpPost(t *testing.T) {
	url := "https://jsonplaceholder.typicode.com/todos"
	header := map[string]string{
		"Content-Type": "application/json",
	}
	type Todo struct {
		UserId int    `json:"userId"`
		Title  string `json:"title"`
	}
	todo := Todo{1, "TestAddToDo"}
	bodyParams, _ := json.Marshal(todo)

	resp, err := HttpPost(url, header, nil, bodyParams)
	if err != nil {
		t.Log("net error: " + err.Error())
		return
	}

	body, _ := io.ReadAll(resp.Body)
	t.Log("response: ", resp.StatusCode, string(body))
}

func TestHttpPostFormData(t *testing.T) {
	apiUrl := "https://jsonplaceholder.typicode.com/todos"
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	postData := url.Values{}
	postData.Add("userId", "1")
	postData.Add("title", "TestToDo")

	resp, err := HttpPost(apiUrl, header, nil, postData)
	if err != nil {
		t.Log("net error: " + err.Error())
		return
	}

	body, _ := io.ReadAll(resp.Body)
	t.Log("response: ", resp.StatusCode, string(body))
}

func TestHttpPut(t *testing.T) {
	url := "https://jsonplaceholder.typicode.com/todos/1"
	header := map[string]string{
		"Content-Type": "application/json",
	}
	type Todo struct {
		Id     int    `json:"id"`
		UserId int    `json:"userId"`
		Title  string `json:"title"`
	}
	todo := Todo{1, 1, "TestPutToDo"}
	bodyParams, _ := json.Marshal(todo)

	resp, err := HttpPut(url, header, nil, bodyParams)
	if err != nil {
		t.Log("net error: ", err.Error())
		return
	}

	body, _ := io.ReadAll(resp.Body)
	t.Log("response: ", resp.StatusCode, string(body))
}

func TestHttpPatch(t *testing.T) {
	url := "https://jsonplaceholder.typicode.com/todos/1"
	header := map[string]string{
		"Content-Type": "application/json",
	}
	type Todo struct {
		Id     int    `json:"id"`
		UserId int    `json:"userId"`
		Title  string `json:"title"`
	}
	todo := Todo{1, 1, "TestPatchToDo"}
	bodyParams, _ := json.Marshal(todo)

	resp, err := HttpPatch(url, header, nil, bodyParams)
	if err != nil {
		t.Log("net error: ", err.Error())
		return
	}

	body, _ := io.ReadAll(resp.Body)
	t.Log("response: ", resp.StatusCode, string(body))
}

func TestHttpDelete(t *testing.T) {
	url := "https://jsonplaceholder.typicode.com/todos/1"
	resp, err := HttpDelete(url)
	if err != nil {
		t.Log("net error: ", err.Error())
		return
	}

	body, _ := io.ReadAll(resp.Body)
	t.Log("response: ", resp.StatusCode, string(body))
}

func TestProxy(t *testing.T) {
	config := &HttpClientConfig{
		HandshakeTimeout: 20 * time.Second,
		ResponseTimeout:  40 * time.Second,
		// Use the proxy ip to add it here
		//Proxy: &url.URL{
		//	Scheme: "http",
		//	Host:   "46.17.63.166:18888",
		//},
	}
	httpClient := NewHttpClientWithConfig(config)
	resp, err := httpClient.Get("https://www.ipplus360.com/getLocation")
	if err != nil {
		t.Log("net error: ", err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
