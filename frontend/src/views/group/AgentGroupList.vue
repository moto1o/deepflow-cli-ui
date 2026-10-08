<template>
  <div>
    <div class="toolbar">
      <el-input
        v-model="filterName"
        placeholder="按名称过滤"
        clearable
        style="width: 240px"
        @keyup.enter="load"
        @clear="load"
      />
      <el-button @click="load">查询</el-button>
      <el-button type="primary" @click="openCreate">创建 Agent 组</el-button>
    </div>

    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="NAME" label="NAME" />
      <el-table-column prop="ID" label="ID (SHORT_UUID)" />
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button type="danger" link @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="createVisible" title="创建 Agent 组" width="480px">
      <el-form :model="createForm" label-width="120px">
        <el-form-item label="名称" required>
          <el-input v-model="createForm.name" placeholder="必填" />
        </el-form-item>
        <el-form-item label="组 ID">
          <el-input v-model="createForm.groupId" placeholder="选填，g- 前缀 + 10 位字母数字" />
          <div class="hint">留空则由 server 自动生成；格式如 g-1yhIguXABC（g- 后接 10 位字母数字）</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { agentGroupApi, type AgentGroup } from '../../api/agentGroup'

const rows = ref<AgentGroup[]>([])
const loading = ref(false)
const filterName = ref('')

const createVisible = ref(false)
const creating = ref(false)
const createForm = reactive({ name: '', groupId: '' })

async function load() {
  loading.value = true
  try {
    rows.value = (await agentGroupApi.list(filterName.value || undefined)) || []
  } catch {
    rows.value = []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  createForm.name = ''
  createForm.groupId = ''
  createVisible.value = true
}

const GROUP_ID_RE = /^g-[0-9a-zA-Z]{10}$/

async function submitCreate() {
  if (!createForm.name) {
    ElMessage.warning('请填写名称')
    return
  }
  if (createForm.groupId && !GROUP_ID_RE.test(createForm.groupId)) {
    ElMessage.warning('组 ID 格式应为 g- 前缀加 10 位字母数字')
    return
  }
  creating.value = true
  try {
    await agentGroupApi.create(createForm.name, createForm.groupId || undefined)
    ElMessage.success('创建成功')
    createVisible.value = false
    await load()
  } finally {
    creating.value = false
  }
}

async function remove(row: AgentGroup) {
  try {
    await ElMessageBox.confirm(`确认删除 Agent 组 "${row.NAME}"？`, '删除确认', {
      type: 'warning',
    })
  } catch {
    return
  }
  await agentGroupApi.remove(row.NAME)
  ElMessage.success('删除成功')
  await load()
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
