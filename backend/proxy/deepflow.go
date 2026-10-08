package proxy

import (
	"fmt"
	"time"
)

// Conn 表示一次请求所使用的 DeepFlow server 连接参数。
// 由前端在每个请求中通过 header 传入，后端不保存状态。
type Conn struct {
	IP      string
	APIPort uint32
	OrgID   int
	Timeout time.Duration
}

func (c Conn) BaseURL() string {
	return fmt.Sprintf("http://%s:%d", c.IP, c.APIPort)
}

func (c Conn) URL(path string) string {
	return c.BaseURL() + path
}
