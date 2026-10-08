import client from './client'

export interface AgentGroup {
  NAME: string
  ID: string
  LCUUID: string
  SHORT_UUID: string
}

export const agentGroupApi = {
  list: (name?: string): Promise<AgentGroup[]> =>
    client.get('/agent-groups', { params: name ? { name } : {} }),
  create: (name: string, groupId?: string) =>
    client.post('/agent-groups', { name, group_id: groupId || '' }),
  remove: (name: string) => client.delete(`/agent-groups/${encodeURIComponent(name)}`),
}
