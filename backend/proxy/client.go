package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const optStatusSuccess = "SUCCESS"

// APIError 表示 DeepFlow server 返回的业务错误（OPT_STATUS != SUCCESS 或非 200）。
type APIError struct {
	StatusCode  int
	Description string
}

func (e *APIError) Error() string {
	return e.Description
}

// Response 是 DeepFlow server 的统一响应结构。
type Response struct {
	OptStatus   string          `json:"OPT_STATUS"`
	Description string          `json:"DESCRIPTION"`
	Data        json.RawMessage `json:"DATA"`
}

// Perform 复刻 CLI 的 CURLPerform：注入统一请求头、解析 OPT_STATUS。
// jsonBody 非 nil 时以 JSON 发送；rawBody 非空时以 x-www-form-urlencoded 发送（用于 YAML）。
func Perform(conn Conn, method, path string, jsonBody map[string]any, rawBody string) (*Response, error) {
	var reader io.Reader
	var contentType string
	if rawBody != "" {
		reader = strings.NewReader(rawBody)
		contentType = "application/x-www-form-urlencoded"
	} else {
		b, _ := json.Marshal(jsonBody)
		reader = bytes.NewReader(b)
		contentType = "application/json"
	}

	req, err := http.NewRequest(method, conn.URL(path), reader)
	if err != nil {
		return nil, err
	}
	if conn.OrgID != 0 {
		req.Header.Set("X-Org-Id", strconv.Itoa(conn.OrgID))
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json, text/plain")
	req.Header.Set("X-User-Id", "1")
	req.Header.Set("X-User-Type", "1")

	client := &http.Client{Timeout: conn.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 %s 失败: %v", req.URL, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 响应失败: %v", req.URL, err)
	}

	var r Response
	if len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, &r); err != nil {
			return nil, fmt.Errorf("解析 %s 响应失败: %v", req.URL, err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		desc := r.Description
		if desc == "" {
			desc = string(respBytes)
		}
		return &r, &APIError{StatusCode: resp.StatusCode, Description: desc}
	}
	if r.OptStatus != "" && r.OptStatus != optStatusSuccess {
		return &r, &APIError{StatusCode: resp.StatusCode, Description: r.Description}
	}
	return &r, nil
}

// DataArray 把 DATA 解析为对象数组。
func (r *Response) DataArray() ([]map[string]any, error) {
	if len(r.Data) == 0 {
		return nil, nil
	}
	var arr []map[string]any
	if err := json.Unmarshal(r.Data, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// DataString 把 DATA 解析为字符串（用于 YAML 模板等）。
func (r *Response) DataString() (string, error) {
	if len(r.Data) == 0 {
		return "", nil
	}
	var s string
	if err := json.Unmarshal(r.Data, &s); err != nil {
		return "", err
	}
	return s, nil
}
