<template>
  <div class="connection-bar">
    <span class="title">DeepFlow Server</span>
    <el-form :inline="true" :model="form" size="default" @submit.prevent>
      <el-form-item label="IP">
        <el-input v-model="form.ip" placeholder="server IP" style="width: 160px" />
      </el-form-item>
      <el-form-item label="API 端口">
        <el-input-number v-model="form.apiPort" :min="1" :max="65535" :controls="false" style="width: 100px" />
      </el-form-item>
      <el-form-item label="Org ID">
        <el-input-number v-model="form.orgId" :min="0" :controls="false" style="width: 80px" />
      </el-form-item>
      <el-form-item label="超时(s)">
        <el-input-number v-model="form.timeout" :min="1" :controls="false" style="width: 80px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="testing" @click="testConnection">测试连接</el-button>
      </el-form-item>
      <el-form-item>
        <el-tag v-if="store.connected" type="success">已连接</el-tag>
        <el-tag v-else type="info">未连接</el-tag>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useConnectionStore } from '../../stores/connection'
import { agentGroupApi } from '../../api/agentGroup'

const store = useConnectionStore()
const form = reactive({ ...store.config })
const testing = ref(false)

async function testConnection() {
  if (!form.ip) {
    ElMessage.warning('请先填写 server IP')
    return
  }
  store.save({ ...form })
  testing.value = true
  try {
    await agentGroupApi.list()
    store.setConnected(true)
    ElMessage.success('连接成功')
  } catch {
    store.setConnected(false)
  } finally {
    testing.value = false
  }
}
</script>

<style scoped>
.connection-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color);
  background: var(--el-bg-color);
}
.connection-bar .title {
  font-weight: 600;
  white-space: nowrap;
}
.connection-bar :deep(.el-form-item) {
  margin-bottom: 0;
}
</style>
