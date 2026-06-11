package rpcserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// NewHandler 创建包含 /rpc 和 /health 的 HTTP handler。
func NewHandler(registry *Registry) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", handleRPC(registry))
	mux.HandleFunc("/health", handleHealth(registry))
	return mux
}

func handleHealth(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"methods": len(registry.Methods()),
		})
	}
}

func handleRPC(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse(nil, CodeInvalidRequest, "method must be POST", nil))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse(nil, CodeParseError, "parse error", nil))
			return
		}
		defer r.Body.Close()

		var raw any
		if err := json.Unmarshal(body, &raw); err != nil {
			writeJSON(w, http.StatusOK, errorResponse(nil, CodeParseError, "parse error", nil))
			return
		}
		if _, ok := raw.([]any); ok {
			writeJSON(w, http.StatusOK, errorResponse(nil, CodeInvalidRequest, "batch requests are not supported", nil))
			return
		}

		var req Request
		if err := json.Unmarshal(body, &req); err != nil {
			writeJSON(w, http.StatusOK, errorResponse(nil, CodeInvalidRequest, "invalid request", nil))
			return
		}
		if !validRequest(req) {
			writeJSON(w, http.StatusOK, errorResponse(req.ID, CodeInvalidRequest, "invalid request", nil))
			return
		}

		result, err := registry.Call(r.Context(), req.Method, req.Params)
		if err != nil {
			var rpcErr *RPCError
			if errors.As(err, &rpcErr) {
				writeJSON(w, http.StatusOK, errorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data))
				return
			}
			writeJSON(w, http.StatusOK, errorResponse(req.ID, CodeInternalError, "internal error", nil))
			return
		}

		writeJSON(w, http.StatusOK, Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		})
	}
}

func validRequest(req Request) bool {
	return req.JSONRPC == "2.0" && len(req.ID) > 0 && req.Method != ""
}

func errorResponse(id json.RawMessage, code int, message string, data any) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
