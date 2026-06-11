package rpcclient

import "context"

type PingParams struct {
	Echo string `json:"echo,omitempty"`
}

type PingResult struct {
	Echo string `json:"echo,omitempty"`
	OK   bool   `json:"ok"`
}

// Ping 调用内置 system.ping 方法。
func (c *Client) Ping(ctx context.Context, echo string) (*PingResult, error) {
	var out PingResult
	err := c.Call(ctx, "system.ping", PingParams{Echo: echo}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Methods 调用内置 system.methods 方法。
func (c *Client) Methods(ctx context.Context) ([]string, error) {
	var out []string
	err := c.Call(ctx, "system.methods", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
