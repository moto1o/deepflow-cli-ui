package handler

import (
	"net/http"

	"github.com/moto1o/deepflow-cli-ui/backend/proxy"
)

// groupLcuuidByName 通过 name 查 vtap-group 的 lcuuid（对照 agent_group.go:141）。
func groupLcuuidByName(conn proxy.Conn, name string) (string, error) {
	resp, err := proxy.Perform(conn, "GET", "/v1/vtap-groups/?name="+name, nil, "")
	if err != nil {
		return "", err
	}
	arr, err := resp.DataArray()
	if err != nil || len(arr) == 0 {
		return "", &proxy.APIError{StatusCode: http.StatusNotFound, Description: "agent-group (" + name + ") 不存在"}
	}
	lcuuid, _ := arr[0]["LCUUID"].(string)
	return lcuuid, nil
}

// ListAgentGroups GET /api/agent-groups
func ListAgentGroups(w http.ResponseWriter, r *http.Request) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	path := "/v1/vtap-groups/"
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
	// 精简为前端需要的字段：NAME、ID(SHORT_UUID)、LCUUID
	out := make([]map[string]any, 0, len(arr))
	for _, g := range arr {
		out = append(out, map[string]any{
			"NAME":       g["NAME"],
			"ID":         g["SHORT_UUID"],
			"LCUUID":     g["LCUUID"],
			"SHORT_UUID": g["SHORT_UUID"],
		})
	}
	writeJSON(w, out)
}

// CreateAgentGroup POST /api/agent-groups  body: {name, group_id?}
func CreateAgentGroup(w http.ResponseWriter, r *http.Request) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		Name    string `json:"name"`
		GroupID string `json:"group_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	_, err = proxy.Perform(conn, "POST", "/v1/vtap-groups/", map[string]any{
		"name":     body.Name,
		"group_id": body.GroupID,
	}, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// DeleteAgentGroup DELETE /api/agent-groups/{name}
func DeleteAgentGroup(w http.ResponseWriter, r *http.Request, name string) {
	conn, err := connFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	lcuuid, err := groupLcuuidByName(conn, name)
	if err != nil {
		writeError(w, err)
		return
	}
	_, err = proxy.Perform(conn, "DELETE", "/v1/vtap-groups/"+lcuuid+"/", nil, "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}
