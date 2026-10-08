import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useConnectionStore } from '../stores/connection'

const client = axios.create({
  baseURL: '/api',
})

// 每次请求注入连接参数到自定义头
client.interceptors.request.use((cfg) => {
  const store = useConnectionStore()
  const c = store.config
  cfg.headers.set('X-Df-Ip', c.ip)
  cfg.headers.set('X-Df-Api-Port', String(c.apiPort))
  cfg.headers.set('X-Df-Org-Id', String(c.orgId))
  cfg.headers.set('X-Df-Timeout', String(c.timeout))
  return cfg
})

// 统一解包 {data} / {error}
client.interceptors.response.use(
  (resp) => resp.data?.data,
  (err) => {
    const msg = err.response?.data?.error || err.message || '请求失败'
    ElMessage.error(msg)
    return Promise.reject(new Error(msg))
  },
)

export default client
