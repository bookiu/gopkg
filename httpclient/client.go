package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bookiu/gopkg/util/types"
	"github.com/google/go-querystring/query"
)

type Client interface {
	Do(ctx context.Context, req *http.Request, result interface{}) error
	Get(ctx context.Context, url string, query interface{}, result interface{}, opts ...RequestOption) error
	Post(ctx context.Context, url string, body interface{}, result interface{}, opts ...RequestOption) error
	PostJson(ctx context.Context, url string, body interface{}, result interface{}, opts ...RequestOption) error
}

type Config struct {
	Timeout   time.Duration
	ProxyFunc func(*http.Request) (*url.URL, error)

	Auth     AuthProvider
	Response ResponseHandler
	Observe  ObserveProvider
}

// RequestOption 定义用于配置请求的函数选项类型
type RequestOption func(*http.Request)

// WithHeader 设置单个请求头
func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

// WithHeaders 设置多个请求头
func WithHeaders(headers map[string]string) RequestOption {
	return func(req *http.Request) {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}
}

// WithContentType 设置 Content-Type 头
func WithContentType(contentType string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set("Content-Type", contentType)
	}
}

// WithBearerToken 设置 Bearer Token 认证头
func WithBearerToken(token string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// WithUserAgent 设置 User-Agent 头
func WithUserAgent(userAgent string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set("User-Agent", userAgent)
	}
}

type HTTPClient struct {
	config *Config
	client *http.Client
}

func NewHTTPClient(config *Config) *HTTPClient {
	if config.Response == nil {
		config.Response = &DirectResponseHandler{} // 默认直接 JSON 解析
	}
	if config.Observe == nil {
		config.Observe = &NoopObserve{}
	}

	return &HTTPClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				Proxy: config.ProxyFunc,
			},
		},
	}
}

func (c *HTTPClient) Do(ctx context.Context, req *http.Request, result interface{}) error {
	if c.config.Auth != nil {
		c.config.Auth.Apply(req)
	}

	startTime := time.Now()
	resp, err := c.client.Do(req.WithContext(ctx))
	c.config.Observe.RecordRequest(ctx, req.Method, req.URL.String(), resp.StatusCode, time.Since(startTime), err)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return c.config.Response.Handle(resp, result)
}

func (c *HTTPClient) Get(ctx context.Context, url string, q interface{}, result interface{}, opts ...RequestOption) error {
	finalUrl := url
	if q != nil {
		v, err := query.Values(q)
		if err != nil {
			return err
		}
		finalUrl = url + "?" + v.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalUrl, nil)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(req)
	}

	return c.Do(ctx, req, result)
}

func (c *HTTPClient) Post(ctx context.Context, url string, body interface{}, result interface{}, opts ...RequestOption) error {
	payload, err := packBody(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req, result)
}

func (c *HTTPClient) PostJson(ctx context.Context, url string, body interface{}, result interface{}, opts ...RequestOption) error {
	opts = append(opts, WithContentType("application/json"))
	return c.Post(ctx, url, body, result, opts...)
}

func packBody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	// if body is struct, return json.NewEncoder(body)
	if types.IsStruct(body) {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		return buf, err
	}
	// if body is io.Reader, return body
	if body, ok := body.(io.Reader); ok {
		return body, nil
	}

	// if body is string, return bytes.NewBufferString(body)
	if body, ok := body.(string); ok {
		return bytes.NewBufferString(body), nil
	}

	// if body is []byte, return bytes.NewBuffer(body)
	if body, ok := body.([]byte); ok {
		return bytes.NewBuffer(body), nil
	}
	return nil, errors.New("unsupported body type")
}
