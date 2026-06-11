package rpcclient

import (
	"encoding/json"
	"fmt"
)

const jsonRPCVersion = "2.0"

// Request 是 JSON-RPC 2.0 单请求格式。
type Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// Response 是 JSON-RPC 2.0 单响应格式。
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError 表示服务端返回的 JSON-RPC 错误对象。
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	if len(e.Data) > 0 {
		return fmt.Sprintf("json-rpc error: code=%d message=%q data=%s", e.Code, e.Message, string(e.Data))
	}
	return fmt.Sprintf("json-rpc error: code=%d message=%q", e.Code, e.Message)
}
