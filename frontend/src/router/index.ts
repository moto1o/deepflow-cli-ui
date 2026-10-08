import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/agents' },
    {
      path: '/agents',
      name: 'agents',
      component: () => import('../views/agent/AgentList.vue'),
    },
    {
      path: '/agent-groups',
      name: 'agent-groups',
      component: () => import('../views/group/AgentGroupList.vue'),
    },
    {
      path: '/agent-group-configs',
      name: 'agent-group-configs',
      component: () => import('../views/config/AgentGroupConfig.vue'),
    },
  ],
})

export default router
