package rpctransport

import (
	"context"
	"encoding/json"

	"github.com/rei0721/go-scaffold/pkg/rpcserver"
)

// NewRegistry 创建应用内置 RPC 方法注册表。
func NewRegistry() (*rpcserver.Registry, error) {
	registry := rpcserver.NewRegistry()
	if err := registry.Register("system.ping", ping); err != nil {
		return nil, err
	}
	if err := registry.Register("system.methods", func(context.Context, json.RawMessage) (any, error) {
		return registry.Methods(), nil
	}); err != nil {
		return nil, err
	}
	return registry, nil
}

func ping(_ context.Context, params json.RawMessage) (any, error) {
	response := map[string]any{"ok": true}
	if len(params) == 0 || string(params) == "null" {
		return response, nil
	}

	var values map[string]any
	if err := json.Unmarshal(params, &values); err != nil {
		return nil, rpcserver.InvalidParams("params must be an object")
	}
	if echo, ok := values["echo"]; ok {
		response["echo"] = echo
	}
	return response, nil
}
