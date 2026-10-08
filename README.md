# DeepFlow CLI Web UI 实现计划（第一期：agent / agent-group / agent-group-config）

## Context

deepflow-ctl 是 Go 编写的 CLI 工具，通过 REST API（默认 `http://{ip}:30417`）管理 DeepFlow server。目标是建一个 Web UI 覆盖 CLI 功能。第一期只做三个最常用模块：**agent、agent-group、agent-group-config**，先交付可用版本，后续再扩展其余模块。

## 技术选型

| 层 | 技术 |
|---|---|
| 前端 | Vite + Vue 3 + TypeScript + Element Plus 2 |
| 状态管理 | Pinia |
| YAML 编辑器 | Monaco Editor |
| 后端代理 | Go + net/http（复用 CLI 的 HTTP 逻辑，避免跨域，统一注入请求头） |

**为什么要后端代理**：浏览器直接调 DeepFlow server 有跨域问题；且 CLI 的多步逻辑（name→lcuuid→操作、YAML 字段转换）用 Go 复刻最省事。后端只做 REST 转发，本期不涉及 gRPC。

## DeepFlow server 连接参数（对照 cli.go:42-47）

| 参数 | 默认值 | 说明 |
|---|---|---|
| ip | 本机默认路由 IP | server 地址 |
| api-port | 30417 | REST API 端口（本期只用这个） |
| rpc-port | 30035 | gRPC 端口（本期不用） |
| org-id | 常量 DEFAULT_ORG_ID（通常 1） | 组织 ID |
| timeout | 30s | 请求超时 |

**请求头（对照 http_util.go:86-92）**：
```
X-User-Id: 1
X-User-Type: 1
X-Org-Id: {orgId}     # 仅当 orgId != 0 时才设置
Accept: application/json, text/plain
Content-Type: application/json           # body 为 JSON 时
Content-Type: application/x-www-form-urlencoded   # body 为原始字符串（如 YAML）时
```

**响应格式**：`{"OPT_STATUS": "SUCCESS", "DATA": ..., "DESCRIPTION": ""}`；`OPT_STATUS != "SUCCESS"` 视为错误，取 DESCRIPTION 作为错误信息。

## 三个模块的精确 API 清单

### 1. agent（agent.go）

| 操作 | 方法 + Endpoint | 说明 |
|---|---|---|
| 列表 | GET `/v1/vtaps/`（可选 `?name=`） | 表格列见下 |
| 删除 | 先 GET `/v1/vtaps/?name={name}` 取 LCUUID，再 DELETE `/v1/vtaps/{lcuuid}/` | 两步 |
| 更新 | 见下方"更新逻辑" | 多步 + 字段转换 |
| 升级列表 | GET `/v1/vtaps/`，过滤 `EXPECTED_REVISION != ""` | 显示 REVISION/EXPECTED_REVISION/UPGRADE_PACKAGE |

**列表表格列**（agent.go:237）：ID, NAME, TYPE, CTRL_IP, CTRL_MAC, STATE, VTAP_GROUP_NAME(显示为 GROUP), EXCEPTIONS, REVISION, UPGRADE_REVISION, POD_CLUSTER_NAME。
- TYPE / STATE / EXCEPTIONS 是枚举整数，需映射为文字（对照 `common/const.go` 的 VtapType / VtapState / VtapException，以及 `trident.Exception_name`）。这些映射表需复制到前端。

**更新逻辑**（agent.go:302-403，Web 表单需复刻）：
- 输入以 agent name 定位，GET `/v1/vtaps/?name=` 取 lcuuid。
- 若含 `tap_mode`：单独 PATCH `/v1/vtaps-tap-mode/`，body `{VTAP_LCUUIDS:[lcuuid], TAP_MODE: <int>}`（tap_mode 名称→int 见 `common.GetVtapTapModeByName`）。
- 若含 `agent_group_id`（或旧字段 `vtap_group_id`）：GET `/v1/vtap-groups/?short_uuid=` 取 group lcuuid，写入 `VTAP_GROUP_LCUUID`。
- 若含 `enable`：转为 `ENABLE`。
- 其余字段连同上述转换结果，PATCH `/v1/vtaps/{lcuuid}/`。
- **前端 UI 简化建议**：编辑弹窗直接提供 分组下拉（复用 agent-group 列表）、启用开关、tap_mode 下拉，把转换逻辑放后端代理做，前端只传结构化字段。

**重平衡**（agent.go:590-676）：
- 检查是否需要：POST `/v1/rebalance-vtap/?check=false&type={controller|analyzer}`，看 `DATA.TOTAL_SWITCH_VTAP_NUM`。
- 执行：POST `/v1/rebalance-vtap/?check=true&type=...`，展示 `DATA.DETAILS`。
- 本期可作为 agent 列表页的一个按钮，优先级低于 CRUD。

### 2. agent-group（agent_group.go）

| 操作 | 方法 + Endpoint | body |
|---|---|---|
| 列表 | GET `/v1/vtap-groups/`（可选 `?name=`） | 列：NAME, SHORT_UUID(显示为 ID) |
| 创建 | POST `/v1/vtap-groups/` | `{name, group_id}`；group_id 选填，规则：`g-` 前缀 + 共 10 位字母数字 |
| 删除 | 先 GET `/v1/vtap-groups/?name={name}` 取 LCUUID，再 DELETE `/v1/vtap-groups/{lcuuid}/` | 两步 |

### 3. agent-group-config（agent_group_config.go）

**核心：配置内容是整段 YAML**，用 Monaco 编辑器。

| 操作 | 方法 + Endpoint | body / 说明 |
|---|---|---|
| 示例模板 | GET `/v1/agent-group-configuration/template/yaml` | 返回 `DATA` 为 YAML 字符串，填入编辑器 |
| 列表(表格) | GET `/v1/vtap-groups/` + GET `/v1/agent-group-configurations`（复数、无尾斜杠），按 AGENT_GROUP_LCUUID join | 显示 NAME, SHORT_UUID |
| 查看单个 YAML | 先由 short_uuid 查 group lcuuid（GET `/v1/vtap-groups/?short_uuid=`），再 GET `/v1/agent-group-configuration/{lcuuid}/yaml` | DATA 为 YAML 字符串 |
| 创建 | 同上取 lcuuid，POST `/v1/agent-group-configuration/{lcuuid}/yaml` | body = 原始 YAML 字符串（Content-Type: x-www-form-urlencoded） |
| 更新 | 同上取 lcuuid，PUT `/v1/agent-group-configuration/{lcuuid}/yaml` | body = 原始 YAML 字符串 |
| 删除 | 同上取 lcuuid，DELETE `/v1/agent-group-configuration/{lcuuid}`（无尾斜杠、无 /yaml） | — |

**注意路径细节**：list 用复数 `agent-group-configurations`，单个操作用单数 `agent-group-configuration`；delete 路径末尾没有 `/yaml` 也没有斜杠。

## 项目结构

```
deepflow-cli-ui/
├── plan.md
├── backend/                       # Go 代理服务
│   ├── go.mod
│   ├── main.go                    # 启动 + 路由注册
│   ├── config.go                  # 读取 server 连接配置（env / 请求头传入）
│   ├── proxy/
│   │   ├── client.go              # 复刻 CURLPerform：注入头、解析 OPT_STATUS
│   │   └── deepflow.go            # 连接参数结构体
│   └── handler/
│       ├── agent.go               # /api/agents 下的 list/delete/update/rebalance
│       ├── agent_group.go         # /api/agent-groups
│       └── agent_group_config.go  # /api/agent-group-configs
└── frontend/                      # Vue 3
    ├── index.html
    ├── vite.config.ts             # dev proxy: /api → backend
    ├── package.json
    ├── tsconfig.json
    └── src/
        ├── main.ts
        ├── App.vue
        ├── router/index.ts
        ├── stores/connection.ts   # server ip/port/orgId，localStorage 持久化
        ├── api/
        │   ├── client.ts          # axios，响应拦截器解析 OPT_STATUS
        │   ├── agent.ts
        │   ├── agentGroup.ts
        │   └── agentGroupConfig.ts
        ├── constants/enums.ts     # VtapType/VtapState/Exception 映射（从 const.go 复制）
        ├── components/
        │   ├── layout/AppLayout.vue
        │   ├── layout/AppSidebar.vue
        │   ├── layout/ConnectionBar.vue   # 顶部：填 ip/port/orgId + 测试连接
        │   └── common/YamlEditor.vue      # Monaco 封装
        └── views/
            ├── agent/AgentList.vue        # 列表 + 删除 + 编辑弹窗 + 升级视图 + 重平衡按钮
            ├── group/AgentGroupList.vue   # 列表 + 创建 + 删除
            └── config/AgentGroupConfig.vue # 列表 + 编辑器(示例/查看/创建/更新/删除)
```

## 代理层 API 设计（前端 → 后端）

后端把 name→lcuuid 的两步查询、YAML 字段转换封装掉，前端调用简洁的 REST：

```
GET    /api/agents                      # 列表（透传 ?name=）
DELETE /api/agents/:name                # 后端内部查 lcuuid 再删
PATCH  /api/agents/:name                # body 传结构化字段，后端做转换
POST   /api/agents/rebalance?type=...&check=...

GET    /api/agent-groups
POST   /api/agent-groups                # {name, group_id?}
DELETE /api/agent-groups/:name

GET    /api/agent-group-configs         # 后端 join vtap-groups + configurations
GET    /api/agent-group-configs/template
GET    /api/agent-group-configs/:shortUuid/yaml
POST   /api/agent-group-configs/:shortUuid/yaml   # body: raw yaml
PUT    /api/agent-group-configs/:shortUuid/yaml   # body: raw yaml
DELETE /api/agent-group-configs/:shortUuid
```

连接参数（ip/api-port/org-id/timeout）由前端每次请求通过自定义头或 query 传给后端，后端据此拼 DeepFlow server URL——避免后端保存状态。

## 实现阶段

### Phase 1：骨架
- Go 后端：main + proxy/client（复刻 CURLPerform 的头注入与 OPT_STATUS 解析）+ 空 handler，跑通 `/api/agents` 转发。
- Vue 前端：Vite + Element Plus + Pinia + Router；AppLayout / AppSidebar / ConnectionBar；connection store 持久化；axios client + 响应拦截器。
- vite dev proxy 把 `/api` 转发到 Go 后端。

### Phase 2：agent-group（最简单，先打通全链路）
- 列表 / 创建（name + 可选 group_id，前端校验 `g-` + 10 位）/ 删除。

### Phase 3：agent
- 列表（枚举映射）、删除、编辑弹窗（分组下拉 + enable 开关 + tap_mode，后端转换）、升级列表视图、重平衡按钮。

### Phase 4：agent-group-config
- YamlEditor 组件；列表（join）；示例模板加载；查看/创建/更新/删除。

## 验证方案

1. 启动：后端 `go run .`（:8080），前端 `npm run dev`（:5173，dev proxy 到 :8080）。
2. ConnectionBar 填入真实 DeepFlow server ip/port/org-id，点"测试连接"（后端 GET `/v1/vtap-groups/` 探活）。
3. 逐模块验证：
   - agent-group：创建 → 列表出现 → 删除消失。
   - agent：列表枚举显示正确 → 编辑分组/启用 → 删除。
   - agent-group-config：加载示例 → 选 group 创建 → 查看 YAML → 更新 → 删除。
4. 异常路径：故意填错 server IP，确认前端能显示后端返回的 DESCRIPTION 错误信息。
```
