package handler

import (
	"net/http"
	"strings"

	"github.com/moto1o/deepflow-cli-ui/backend/proxy"
)

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}

func agentLcuuidByName(conn proxy.Conn, name string) (map[string]any, error) {
	resp, err := proxy.Perform(conn, "GET", "/v1/vtaps/?name="+name, nil, "")
	if err != nil {
		return nil, err
	}
	arr, err := resp.DataArray()
	if err != nil || len(arr) == 0 {
		return nil, &proxy.APIError{StatusCode: http.StatusNotFound, Description: "agent (" + name + ") 不存在"}
	}
	return arr[0], nil
}

// ListAgents GET /api/agents  可选 ?name= 和 ?upgrade=true（升级视图）
func containsIgnoreCase(s, sub string) bool {
	if sub == "" {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

// ListAgents GET /api/agents  可选 ?name= ?group= ?pod_cluster= ?upgrade=true
func ListAgents(w http.ResponseWriter, r *http.Request) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	path := "/v1/vtaps/"
	if name := r.URL.Query().Get("name"); name != "" {
		path += "?name=" + name
	}
	resp, err := proxy.Perform(conn, "GET", path, nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	arr, err := resp.DataArray()
	if err != nil {
		writeError(w, err)
		return
	}

	filterGroup := r.URL.Query().Get("group")
	filterPodCluster := r.URL.Query().Get("pod_cluster")

	upgradeView := r.URL.Query().Get("upgrade") == "true"
	out := make([]map[string]any, 0, len(arr))
	for _, v := range arr {
		if upgradeView {
			exp, _ := v["EXPECTED_REVISION"].(string)
			if exp == "" {
				continue
			}
			revision, _ := v["REVISION"].(string)
			complete, _ := v["COMPLETE_REVISION"].(string)
			out = append(out, map[string]any{
				"NAME":              v["NAME"],
				"REVISION":          revision + "-" + complete,
				"EXPECTED_REVISION": exp,
				"UPGRADE_PACKAGE":   v["UPGRADE_PACKAGE"],
			})
			continue
		}

		groupName, _ := v["VTAP_GROUP_NAME"].(string)
		podCluster, _ := v["POD_CLUSTER_NAME"].(string)
		if !containsIgnoreCase(groupName, filterGroup) {
			continue
		}
		if !containsIgnoreCase(podCluster, filterPodCluster) {
			continue
		}

		exceptions := []string{}
		if raw, ok := v["EXCEPTIONS"].([]any); ok {
			for _, e := range raw {
				if s := vtapExceptionStr(toInt64(e)); s != "" {
					exceptions = append(exceptions, s)
				}
			}
		}
		out = append(out, map[string]any{
			"ID":               v["ID"],
			"NAME":             v["NAME"],
			"TYPE":             vtapTypeStr(toInt(v["TYPE"])),
			"CTRL_IP":          v["CTRL_IP"],
			"CTRL_MAC":         v["CTRL_MAC"],
			"STATE":            vtapStateStr(toInt(v["STATE"])),
			"GROUP":            groupName,
			"EXCEPTIONS":       exceptions,
			"REVISION":         v["REVISION"],
			"UPGRADE_REVISION": v["UPGRADE_REVISION"],
			"POD_CLUSTER":      podCluster,
			"LCUUID":           v["LCUUID"],
		})
	}
	writeJSON(w, out)
}

// DeleteAgent DELETE /api/agents/{name}
func DeleteAgent(w http.ResponseWriter, r *http.Request, name string) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	agent, err := agentLcuuidByName(conn, name)
	if err != nil {
		writeError(w, err)
		return
	}
	lcuuid, _ := agent["LCUUID"].(string)
	_, err = proxy.Perform(conn, "DELETE", "/v1/vtaps/"+lcuuid+"/", nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// UpdateAgent PATCH /api/agents/{name}
// body: { enable?: bool, agent_group_id?: string, tap_mode?: string, ...其余透传 }
// 复刻 agent.go 的字段转换逻辑。
func UpdateAgent(w http.ResponseWriter, r *http.Request, name string) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body map[string]any
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}

	agent, err := agentLcuuidByName(conn, name)
	if err != nil {
		writeError(w, err)
		return
	}
	lcuuid, _ := agent["LCUUID"].(string)

	// tap_mode：单独走 /v1/vtaps-tap-mode/
	if tm, ok := body["tap_mode"].(string); ok && tm != "" {
		mode := tapModeValue[tm]
		_, err = proxy.Perform(conn, "PATCH", "/v1/vtaps-tap-mode/", map[string]any{
			"VTAP_LCUUIDS": []string{lcuuid},
			"TAP_MODE":     mode,
		}, "")
		if err != nil {
			writeError(w, err)
			return
		}
		delete(body, "tap_mode")
	}

	// agent_group_id → 查 group lcuuid → VTAP_GROUP_LCUUID
	if gid, ok := body["agent_group_id"].(string); ok && gid != "" {
		resp, err := proxy.Perform(conn, "GET", "/v1/vtap-groups/?short_uuid="+gid, nil, "")
		if err != nil {
			writeError(w, err)
			return
		}
		garr, _ := resp.DataArray()
		if len(garr) == 0 {
			writeError(w, &proxy.APIError{StatusCode: http.StatusNotFound, Description: "agent-group (" + gid + ") 不存在"})
			return
		}
		body["VTAP_GROUP_LCUUID"], _ = garr[0]["LCUUID"].(string)
		delete(body, "agent_group_id")
	}

	// enable → ENABLE（DeepFlow server 期望 int：1=启用 0=禁用）
	if en, ok := body["enable"]; ok {
		if b, ok2 := en.(bool); ok2 {
			if b {
				body["ENABLE"] = 1
			} else {
				body["ENABLE"] = 0
			}
		} else {
			body["ENABLE"] = en
		}
		delete(body, "enable")
	}

	if len(body) == 0 {
		writeJSON(w, map[string]any{"ok": true})
		return
	}

	_, err = proxy.Perform(conn, "PATCH", "/v1/vtaps/"+lcuuid+"/", body, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// RebalanceAgent POST /api/agents/rebalance?type=controller|analyzer&check=true|false
func RebalanceAgent(w http.ResponseWriter, r *http.Request) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	typ := r.URL.Query().Get("type")
	if typ == "" {
		typ = "controller"
	}
	check := r.URL.Query().Get("check")
	if check == "" {
		check = "false"
	}
	path := "/v1/rebalance-vtap/?check=" + check + "&type=" + typ
	resp, err := proxy.Perform(conn, "POST", path, nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, resp.Data)
}
