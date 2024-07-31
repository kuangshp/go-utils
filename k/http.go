package k

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// HttpGet send get http request.
func HttpGet(url string, params ...any) (*http.Response, error) {
	return doHttpRequest(http.MethodGet, url, params...)
}

// HttpPost send post http request.
func HttpPost(url string, params ...any) (*http.Response, error) {
	return doHttpRequest(http.MethodPost, url, params...)
}

// HttpPut send put http request.
func HttpPut(url string, params ...any) (*http.Response, error) {
	return doHttpRequest(http.MethodPut, url, params...)
}

// HttpDelete send delete http request.
func HttpDelete(url string, params ...any) (*http.Response, error) {
	return doHttpRequest(http.MethodDelete, url, params...)
}

// HttpPatch send patch http request.
func HttpPatch(url string, params ...any) (*http.Response, error) {
	return doHttpRequest(http.MethodPatch, url, params...)
}

// ParseHttpResponse decode http response to specified interface.
func ParseHttpResponse(resp *http.Response, obj any) error {
	if resp == nil {
		return errors.New("InvalidResp")
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

// ConvertMapToQueryString convert map to sorted url query string.
// Play: https://go.dev/play/p/jnNt_qoSnRi
func ConvertMapToQueryString(param map[string]any) string {
	if param == nil {
		return ""
	}
	var keys []string
	for key := range param {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var build strings.Builder
	for i, v := range keys {
		build.WriteString(v)
		build.WriteString("=")
		build.WriteString(fmt.Sprintf("%v", param[v]))
		if i != len(keys)-1 {
			build.WriteString("&")
		}
	}
	return build.String()
}

// HttpRequest struct is a composed http request
type HttpRequest struct {
	RawURL      string
	Method      string
	Headers     http.Header
	QueryParams url.Values
	FormData    url.Values
	File        *File
	Body        []byte
}

// HttpClientConfig contains some configurations for http client
type HttpClientConfig struct {
	Timeout          time.Duration
	SSLEnabled       bool
	TLSConfig        *tls.Config
	Compressed       bool
	HandshakeTimeout time.Duration
	ResponseTimeout  time.Duration
	Verbose          bool
	Proxy            *url.URL
}

// defaultHttpClientConfig defalut client config.
var defaultHttpClientConfig = &HttpClientConfig{
	Timeout:          50 * time.Second,
	Compressed:       false,
	HandshakeTimeout: 10 * time.Second,
	ResponseTimeout:  10 * time.Second,
}

// HttpClient is used for sending http request.
type HttpClient struct {
	*http.Client
	TLS     *tls.Config
	Request *http.Request
	Config  HttpClientConfig
	Context context.Context
}

// NewHttpClient make a HttpClient instance.
func NewHttpClient() *HttpClient {
	client := &HttpClient{
		Client: &http.Client{
			Timeout: defaultHttpClientConfig.Timeout,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   defaultHttpClientConfig.HandshakeTimeout,
				ResponseHeaderTimeout: defaultHttpClientConfig.ResponseTimeout,
				DisableCompression:    !defaultHttpClientConfig.Compressed,
			},
		},
		Config: *defaultHttpClientConfig,
	}

	return client
}

// NewHttpClientWithConfig make a HttpClient instance with pass config.
func NewHttpClientWithConfig(config *HttpClientConfig) *HttpClient {
	if config == nil {
		config = defaultHttpClientConfig
	}

	client := &HttpClient{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSHandshakeTimeout:   config.HandshakeTimeout,
				ResponseHeaderTimeout: config.ResponseTimeout,
				DisableCompression:    !config.Compressed,
			},
		},
		Config: *config,
	}

	if config.SSLEnabled {
		client.TLS = config.TLSConfig
	}

	if config.Proxy != nil {
		transport := client.Client.Transport.(*http.Transport)
		transport.Proxy = http.ProxyURL(config.Proxy)
	}

	return client
}

// SendRequest send http request.
// Play: https://go.dev/play/p/jUSgynekH7G
func (client *HttpClient) SendRequest(request *HttpRequest) (*http.Response, error) {
	err := validateRequest(request)
	if err != nil {
		return nil, err
	}

	rawUrl := request.RawURL

	req, err := http.NewRequest(request.Method, rawUrl, bytes.NewBuffer(request.Body))

	if client.Context != nil {
		req, err = http.NewRequestWithContext(client.Context, request.Method, rawUrl, bytes.NewBuffer(request.Body))
	}

	if err != nil {
		return nil, err
	}

	client.setTLS(rawUrl)
	client.setHeader(req, request.Headers)

	err = client.setQueryParam(req, rawUrl, request.QueryParams)
	if err != nil {
		return nil, err
	}

	if request.FormData != nil {
		if request.File != nil {
			err = client.setFormData(req, request.FormData, setFile(request.File))
		} else {
			err = client.setFormData(req, request.FormData, nil)
		}
	}

	client.Request = req

	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DecodeResponse decode response into target object.
// Play: https://go.dev/play/p/jUSgynekH7G
func (client *HttpClient) DecodeResponse(resp *http.Response, target any) error {
	if resp == nil {
		return errors.New("invalid target param")
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// setTLS set http client transport TLSClientConfig
func (client *HttpClient) setTLS(rawUrl string) {
	if strings.HasPrefix(rawUrl, "https") {
		if transport, ok := client.Client.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = client.TLS
		}
	}
}

// setHeader set http request header
func (client *HttpClient) setHeader(req *http.Request, headers http.Header) {
	if headers == nil {
		headers = make(http.Header)
	}

	if _, ok := headers["Accept"]; !ok {
		headers["Accept"] = []string{"*/*"}
	}
	if _, ok := headers["Accept-Encoding"]; !ok && client.Config.Compressed {
		headers["Accept-Encoding"] = []string{"deflate, gzip"}
	}

	req.Header = headers
}

// setQueryParam set http request query string param
func (client *HttpClient) setQueryParam(req *http.Request, reqUrl string, queryParam url.Values) error {
	if queryParam != nil {
		if !strings.Contains(reqUrl, "?") {
			reqUrl = reqUrl + "?" + queryParam.Encode()
		} else {
			reqUrl = reqUrl + "&" + queryParam.Encode()
		}
		u, err := url.Parse(reqUrl)
		if err != nil {
			return err
		}
		req.URL = u
	}
	return nil
}

// setFormData set http request FormData param
func (client *HttpClient) setFormData(req *http.Request, values url.Values, setFile SetFileFunc) error {
	if setFile != nil {
		err := setFile(req, values)
		if err != nil {
			return err
		}
	} else {
		formData := []byte(values.Encode())
		req.Body = io.NopCloser(bytes.NewReader(formData))
		req.ContentLength = int64(len(formData))
	}
	return nil
}

type SetFileFunc func(req *http.Request, values url.Values) error

// File struct is a combination of file attributes
type File struct {
	Content   []byte
	Path      string
	FieldName string
	FileName  string
}

// setFile set parameters for http request formdata file upload
func setFile(f *File) SetFileFunc {
	return func(req *http.Request, values url.Values) error {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		for key, vals := range values {
			for _, val := range vals {
				err := writer.WriteField(key, val)
				if err != nil {
					return err
				}
			}
		}

		if f.Content != nil {
			part, err := writer.CreateFormFile(f.FieldName, f.FileName)
			if err != nil {
				return err
			}
			part.Write(f.Content)
		} else if f.Path != "" {
			file, err := os.Open(f.Path)
			if err != nil {
				return err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(f.FieldName, f.FileName)
			if err != nil {
				return err
			}
			_, err = io.Copy(part, file)
			if err != nil {
				return err
			}
		}

		err := writer.Close()
		if err != nil {
			return err
		}

		req.Body = io.NopCloser(body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.ContentLength = int64(body.Len())

		return nil
	}
}

// validateRequest check if a request has url, and valid method.
func validateRequest(req *HttpRequest) error {
	if req.RawURL == "" {
		return errors.New("invalid request url")
	}
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH",
		"HEAD", "CONNECT", "OPTIONS", "TRACE"}

	if !IsContains(methods, strings.ToUpper(req.Method)) {
		return errors.New("invalid request method")
	}

	return nil
}

// StructToUrlValues convert struct to url valuse,
// only convert the field which is exported and has `json` tag.
// Play: https://go.dev/play/p/pFqMkM40w9z
func StructToUrlValues(targetStruct any) (url.Values, error) {
	result := url.Values{}

	var m map[string]interface{}

	jsonBytes, err := json.Marshal(targetStruct)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &m)
	if err != nil {
		return nil, err
	}

	for k, v := range m {
		result.Add(k, fmt.Sprintf("%v", v))
	}

	return result, nil
}

func doHttpRequest(method, reqUrl string, params ...any) (*http.Response, error) {
	if len(reqUrl) == 0 {
		return nil, errors.New("url should be specified")
	}

	req := &http.Request{
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	client := &http.Client{}
	err := setUrl(req, reqUrl)
	if err != nil {
		return nil, err
	}

	switch len(params) {
	case 1:
		err = setHeader(req, params[0])
		if err != nil {
			return nil, err
		}
	case 2:
		err := setHeaderAndQueryParam(req, reqUrl, params[0], params[1])
		if err != nil {
			return nil, err
		}
	case 3:
		err := setHeaderAndQueryAndBody(req, reqUrl, params[0], params[1], params[2])
		if err != nil {
			return nil, err
		}
	case 4:
		err := setHeaderAndQueryAndBody(req, reqUrl, params[0], params[1], params[2])
		if err != nil {
			return nil, err
		}
		client, err = getClient(params[3])
		if err != nil {
			return nil, err
		}
	}

	resp, e := client.Do(req)
	return resp, e
}

func setHeaderAndQueryParam(req *http.Request, reqUrl string, header, queryParam any) error {
	err := setHeader(req, header)
	if err != nil {
		return err
	}
	err = setQueryParam(req, reqUrl, queryParam)
	if err != nil {
		return err
	}
	return nil
}

func setHeaderAndQueryAndBody(req *http.Request, reqUrl string, header, queryParam, body any) error {
	if err := setHeader(req, header); err != nil {
		return err
	} else if err = setQueryParam(req, reqUrl, queryParam); err != nil {
		return err
	} else if err = setBodyByte(req, body); err != nil {
		return err
	}
	return nil
}

func setHeader(req *http.Request, header any) error {
	if header == nil {
		return nil
	}

	switch v := header.(type) {
	case map[string]string:
		for k := range v {
			req.Header.Add(k, v[k])
		}
	case http.Header:
		for k, vv := range v {
			for _, vvv := range vv {
				req.Header.Add(k, vvv)
			}
		}
	default:
		return errors.New("header params type should be http.Header or map[string]string")
	}

	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	return nil
}

func setUrl(req *http.Request, reqUrl string) error {
	u, err := url.Parse(reqUrl)
	if err != nil {
		return err
	}
	req.URL = u
	return nil
}

func setQueryParam(req *http.Request, reqUrl string, queryParam any) error {
	if queryParam == nil {
		return nil
	}

	var values url.Values
	switch v := queryParam.(type) {
	case map[string]string:
		values = url.Values{}
		for k := range v {
			values.Set(k, v[k])
		}
	case url.Values:
		values = v
	default:
		return errors.New("query string params type should be url.Values or map[string]string")
	}

	// set url
	if values != nil {
		if !strings.Contains(reqUrl, "?") {
			reqUrl = reqUrl + "?" + values.Encode()
		} else {
			reqUrl = reqUrl + "&" + values.Encode()
		}
	}
	u, err := url.Parse(reqUrl)
	if err != nil {
		return err
	}
	req.URL = u

	return nil
}

func setBodyByte(req *http.Request, body any) error {
	if body == nil {
		return nil
	}
	var bodyReader *bytes.Reader
	switch b := body.(type) {
	case io.Reader:
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, b); err != nil {
			return err
		}
		req.Body = ioutil.NopCloser(buf)
		req.ContentLength = int64(buf.Len())
	case []byte:
		bodyReader = bytes.NewReader(b)
		req.Body = ioutil.NopCloser(bodyReader)
		req.ContentLength = int64(bodyReader.Len())
	case map[string]interface{}:
		values := url.Values{}
		for k := range b {
			values.Set(k, fmt.Sprintf("%v", b[k]))
		}
		bodyReader = bytes.NewReader([]byte(values.Encode()))
		req.Body = ioutil.NopCloser(bodyReader)
		req.ContentLength = int64(bodyReader.Len())
	case url.Values:
		bodyReader = bytes.NewReader([]byte(b.Encode()))
		req.Body = ioutil.NopCloser(bodyReader)
		req.ContentLength = int64(bodyReader.Len())
	default:
		return fmt.Errorf("invalid body type: %T", b)
	}
	return nil
}

func getClient(client any) (*http.Client, error) {
	c := http.Client{}
	if client != nil {
		switch v := client.(type) {
		case http.Client:
			c = v
		default:
			return nil, errors.New("client type should be http.Client")
		}
	}

	return &c, nil
}
