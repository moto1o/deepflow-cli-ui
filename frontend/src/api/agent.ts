import client from './client'

export interface Agent {
  ID: number
  NAME: string
  TYPE: string
  CTRL_IP: string
  CTRL_MAC: string
  STATE: string
  GROUP: string
  EXCEPTIONS: string[]
  REVISION: string
  UPGRADE_REVISION: string
  POD_CLUSTER: string
  LCUUID: string
}

export interface AgentUpgrade {
  NAME: string
  REVISION: string
  EXPECTED_REVISION: string
  UPGRADE_PACKAGE: string
}

export interface AgentUpdatePayload {
  enable?: boolean
  agent_group_id?: string
  tap_mode?: string
}

export interface AgentListParams {
  name?: string
  group?: string
  pod_cluster?: string
}

export const agentApi = {
  list: (params?: AgentListParams): Promise<Agent[]> =>
    client.get('/agents', { params }),
  listUpgrade: (): Promise<AgentUpgrade[]> =>
    client.get('/agents', { params: { upgrade: 'true' } }),
  update: (name: string, payload: AgentUpdatePayload) =>
    client.patch(`/agents/${encodeURIComponent(name)}`, payload),
  remove: (name: string) => client.delete(`/agents/${encodeURIComponent(name)}`),
  rebalance: (type: string, check: boolean) =>
    client.post('/agents/rebalance', null, { params: { type, check: String(check) } }),
}
