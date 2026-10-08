package handler

import (
	"net/http"

	"github.com/moto1o/deepflow-cli-ui/backend/proxy"
)

// groupLcuuidByShortUUID 通过 short_uuid 查 group lcuuid（对照 agent_group_config.go:207）。
func groupLcuuidByShortUUID(conn proxy.Conn, shortUUID string) (string, error) {
	resp, err := proxy.Perform(conn, "GET", "/v1/vtap-groups/?short_uuid="+shortUUID, nil, "")
	if err != nil {
		return "", err
	}
	arr, err := resp.DataArray()
	if err != nil || len(arr) == 0 {
		return "", &proxy.APIError{StatusCode: http.StatusNotFound, Description: "agent-group (" + shortUUID + ") 不存在"}
	}
	lcuuid, _ := arr[0]["LCUUID"].(string)
	return lcuuid, nil
}

// GetConfigTemplate GET /api/agent-group-configs/template
func GetConfigTemplate(w http.ResponseWriter, r *http.Request) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	resp, err := proxy.Perform(conn, "GET", "/v1/agent-group-configuration/template/yaml", nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	yaml, err := resp.DataString()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"yaml": yaml})
}

// ListConfigs GET /api/agent-group-configs
// join /v1/vtap-groups/ 与 /v1/agent-group-configurations（对照 listAgentGroupConfig 表格分支）。
func ListConfigs(w http.ResponseWriter, r *http.Request) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	groupResp, err := proxy.Perform(conn, "GET", "/v1/vtap-groups/", nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	groups, _ := groupResp.DataArray()
	type info struct{ name, shortUUID string }
	byLcuuid := map[string]info{}
	for _, g := range groups {
		lcuuid, _ := g["LCUUID"].(string)
		name, _ := g["NAME"].(string)
		su, _ := g["SHORT_UUID"].(string)
		byLcuuid[lcuuid] = info{name, su}
	}

	cfgResp, err := proxy.Perform(conn, "GET", "/v1/agent-group-configurations", nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	cfgs, _ := cfgResp.DataArray()
	out := make([]map[string]any, 0, len(cfgs))
	for _, c := range cfgs {
		lcuuid, _ := c["AGENT_GROUP_LCUUID"].(string)
		if i, ok := byLcuuid[lcuuid]; ok {
			out = append(out, map[string]any{
				"NAME":           i.name,
				"AGENT_GROUP_ID": i.shortUUID,
			})
		}
	}
	writeJSON(w, out)
}

// GetConfigYAML GET /api/agent-group-configs/{shortUuid}/yaml
func GetConfigYAML(w http.ResponseWriter, r *http.Request, shortUUID string) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	lcuuid, err := groupLcuuidByShortUUID(conn, shortUUID)
	if err != nil {
		writeError(w, err)
		return
	}
	resp, err := proxy.Perform(conn, "GET", "/v1/agent-group-configuration/"+lcuuid+"/yaml", nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	yaml, err := resp.DataString()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"yaml": yaml})
}

// SaveConfigYAML 处理创建(POST)与更新(PUT)。body 为原始 YAML。
func SaveConfigYAML(w http.ResponseWriter, r *http.Request, shortUUID, method string) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	lcuuid, err := groupLcuuidByShortUUID(conn, shortUUID)
	if err != nil {
		writeError(w, err)
		return
	}
	yaml := readRawBody(r)
	_, err = proxy.Perform(conn, method, "/v1/agent-group-configuration/"+lcuuid+"/yaml", nil, yaml)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// DeleteConfig DELETE /api/agent-group-configs/{shortUuid}
func DeleteConfig(w http.ResponseWriter, r *http.Request, shortUUID string) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	lcuuid, err := groupLcuuidByShortUUID(conn, shortUUID)
	if err != nil {
		writeError(w, err)
		return
	}
	// 注意：末尾无 /yaml、无斜杠（对照 agent_group_config.go:286）
	_, err = proxy.Perform(conn, "DELETE", "/v1/agent-group-configuration/"+lcuuid, nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}
