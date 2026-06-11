package rpcclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// Client 封装 demo 插件访问主服务 RPC 入口所需的通用能力。
type Client struct {
	baseURL    string
	rpcURL     string
	healthURL  string
	httpClient *http.Client
	nextID     atomic.Int64
}

type Option func(*Client)

// WithHTTPClient 允许调用方注入自定义 HTTP 客户端。
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// New 创建 RPC 客户端。
func New(baseURL string, opts ...Option) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return nil, errors.New("baseURL is required")
	}

	c := &Client{
		baseURL:   baseURL,
		rpcURL:    baseURL + "/rpc",
		healthURL: baseURL + "/health",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Health 调用 RPC 端口的存活检查。
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.healthURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("health check failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// Call 发送一个 JSON-RPC 2.0 单请求。
func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	if method == "" {
		return errors.New("method is required")
	}

	id := c.nextID.Add(1)
	payload := Request{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Method:  method,
		Params:  params,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal json-rpc request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("read json-rpc response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rpc http error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var rpcResp Response
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return fmt.Errorf("decode json-rpc response: %w; body=%s", err, string(respBody))
	}
	if rpcResp.JSONRPC != jsonRPCVersion {
		return fmt.Errorf("invalid json-rpc version: %q", rpcResp.JSONRPC)
	}
	if rpcResp.ID != id {
		return fmt.Errorf("json-rpc id mismatch: want=%d got=%d", id, rpcResp.ID)
	}
	if rpcResp.Error != nil {
		return rpcResp.Error
	}
	if result == nil {
		return nil
	}
	if len(rpcResp.Result) == 0 {
		return errors.New("json-rpc result is empty")
	}
	if err := json.Unmarshal(rpcResp.Result, result); err != nil {
		return fmt.Errorf("decode json-rpc result: %w", err)
	}

	return nil
}
