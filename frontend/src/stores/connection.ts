import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORAGE_KEY = 'deepflow-connection'

export interface ConnectionConfig {
  ip: string
  apiPort: number
  orgId: number
  timeout: number
}

function load(): ConnectionConfig {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (raw) {
    try {
      return JSON.parse(raw)
    } catch {
      // ignore
    }
  }
  return { ip: '', apiPort: 30417, orgId: 1, timeout: 30 }
}

export const useConnectionStore = defineStore('connection', () => {
  const config = ref<ConnectionConfig>(load())
  const connected = ref(false)

  function save(next: ConnectionConfig) {
    config.value = { ...next }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(config.value))
  }

  function setConnected(v: boolean) {
    connected.value = v
  }

  return { config, connected, save, setConnected }
})
