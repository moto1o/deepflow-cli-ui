import client from './client'

export interface AgentGroupConfig {
  NAME: string
  AGENT_GROUP_ID: string
}

export const agentGroupConfigApi = {
  list: (): Promise<AgentGroupConfig[]> => client.get('/agent-group-configs'),
  template: (): Promise<{ yaml: string }> =>
    client.get('/agent-group-configs/template'),
  getYaml: (shortUuid: string): Promise<{ yaml: string }> =>
    client.get(`/agent-group-configs/${encodeURIComponent(shortUuid)}/yaml`),
  create: (shortUuid: string, yaml: string) =>
    client.post(`/agent-group-configs/${encodeURIComponent(shortUuid)}/yaml`, yaml, {
      headers: { 'Content-Type': 'text/plain' },
    }),
  update: (shortUuid: string, yaml: string) =>
    client.put(`/agent-group-configs/${encodeURIComponent(shortUuid)}/yaml`, yaml, {
      headers: { 'Content-Type': 'text/plain' },
    }),
  remove: (shortUuid: string) =>
    client.delete(`/agent-group-configs/${encodeURIComponent(shortUuid)}`),
}
