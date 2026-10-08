package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/moto1o/deepflow-cli-ui/backend/proxy"
)

// connFromRequest 从请求头读取 DeepFlow server 连接参数。
// 前端每次请求都带上，后端不保存状态。
func connFromRequest(r *http.Request) (proxy.Conn, error) {
	ip := r.Header.Get("X-Df-Ip")
	if ip == "" {
		return proxy.Conn{}, errors.New("缺少 server 地址（X-Df-Ip）")
	}
	apiPort := parseUint(r.Header.Get("X-Df-Api-Port"), 30417)
	orgID := int(parseUint(r.Header.Get("X-Df-Org-Id"), 1))
	timeout := parseUint(r.Header.Get("X-Df-Timeout"), 30)
	return proxy.Conn{
		IP:      ip,
		APIPort: apiPort,
		OrgID:   orgID,
		Timeout: time.Duration(timeout) * time.Second,
	}, nil
}

func parseUint(s string, def uint32) uint32 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return def
	}
	return uint32(v)
}

// writeJSON 返回成功数据。
func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": data})
}

// writeError 返回错误。业务错误取 DeepFlow 的 DESCRIPTION。
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	var apiErr *proxy.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode != 0 {
		status = apiErr.StatusCode
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
}

// readRawBody 读取请求体为字符串（用于 YAML 透传）。
func readRawBody(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	b, _ := io.ReadAll(r.Body)
	return string(b)
}

// decodeJSON 解析请求体 JSON。
func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
