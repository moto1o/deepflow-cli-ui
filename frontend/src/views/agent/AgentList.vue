<template>
  <div>
    <div class="toolbar">
      <el-radio-group v-model="view" @change="onViewChange">
        <el-radio-button label="list">列表</el-radio-button>
        <el-radio-button label="upgrade">升级</el-radio-button>
      </el-radio-group>
      <template v-if="view === 'list'">
        <el-input
          v-model="filterName"
          placeholder="NAME"
          clearable
          style="width: 180px"
          @keyup.enter="load"
          @clear="load"
        />
        <el-select
          v-model="filterGroup"
          clearable
          placeholder="GROUP"
          style="width: 200px"
          @change="load"
          @clear="load"
        >
          <el-option
            v-for="g in groups"
            :key="g.NAME"
            :label="g.NAME"
            :value="g.NAME"
          />
        </el-select>
        <el-input
          v-model="filterPodCluster"
          placeholder="POD_CLUSTER"
          clearable
          style="width: 180px"
          @keyup.enter="load"
          @clear="load"
        />
        <el-button @click="load">查询</el-button>
      </template>
      <el-button @click="rebalance('controller')">控制器重平衡</el-button>
      <el-button @click="rebalance('analyzer')">数据节点重平衡</el-button>
    </div>

    <!-- 列表视图 -->
    <el-table v-if="view === 'list'" :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="ID" label="ID" width="70" />
      <el-table-column prop="NAME" label="NAME" min-width="140" />
      <el-table-column prop="TYPE" label="TYPE" width="110" />
      <el-table-column prop="CTRL_IP" label="CTRL_IP" width="130" />
      <el-table-column prop="CTRL_MAC" label="CTRL_MAC" width="150" />
      <el-table-column prop="STATE" label="STATE" width="90" />
      <el-table-column prop="GROUP" label="GROUP" min-width="120" />
      <el-table-column label="EXCEPTIONS" min-width="140">
        <template #default="{ row }">
          <el-tag v-for="e in row.EXCEPTIONS" :key="e" type="danger" size="small" class="ex-tag">{{ e }}</el-tag>
          <span v-if="!row.EXCEPTIONS || !row.EXCEPTIONS.length">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="REVISION" label="REVISION" min-width="120" />
      <el-table-column prop="UPGRADE_REVISION" label="UPGRADE_REVISION" min-width="140" />
      <el-table-column prop="POD_CLUSTER" label="POD_CLUSTER" min-width="120" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="openEdit(row)">编辑</el-button>
          <el-button type="danger" link @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 升级视图 -->
    <el-table v-else :data="upgradeRows" v-loading="loading" border stripe>
      <el-table-column prop="NAME" label="NAME" min-width="160" />
      <el-table-column prop="REVISION" label="REVISION" min-width="160" />
      <el-table-column prop="EXPECTED_REVISION" label="EXPECTED_REVISION" min-width="160" />
      <el-table-column prop="UPGRADE_PACKAGE" label="UPGRADE_PACKAGE" min-width="200" />
    </el-table>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editVisible" :title="`编辑 Agent: ${editName}`" width="480px">
      <el-form :model="editForm" label-width="120px">
        <el-form-item label="启用">
          <el-switch v-model="editForm.enable" />
        </el-form-item>
        <el-form-item label="Agent 组">
          <el-select v-model="editForm.agent_group_id" clearable placeholder="不修改则留空" style="width: 100%">
            <el-option
              v-for="g in groups"
              :key="g.SHORT_UUID"
              :label="`${g.NAME} (${g.SHORT_UUID})`"
              :value="g.SHORT_UUID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="采集模式">
          <el-select v-model="editForm.tap_mode" clearable placeholder="不修改则留空" style="width: 100%">
            <el-option v-for="t in TAP_MODE_OPTIONS" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { agentApi, type Agent, type AgentListParams, type AgentUpgrade, type AgentUpdatePayload } from '../../api/agent'
import { agentGroupApi, type AgentGroup } from '../../api/agentGroup'
import { TAP_MODE_OPTIONS } from '../../constants/enums'

const view = ref<'list' | 'upgrade'>('list')
const rows = ref<Agent[]>([])
const upgradeRows = ref<AgentUpgrade[]>([])
const groups = ref<AgentGroup[]>([])
const loading = ref(false)
const filterName = ref('')
const filterGroup = ref('')
const filterPodCluster = ref('')

const editVisible = ref(false)
const saving = ref(false)
const editName = ref('')
const editForm = reactive<{ enable: boolean; agent_group_id: string; tap_mode: string }>({
  enable: true,
  agent_group_id: '',
  tap_mode: '',
})

async function load() {
  loading.value = true
  try {
    const params: AgentListParams = {}
    if (filterName.value) params.name = filterName.value
    if (filterGroup.value) params.group = filterGroup.value
    if (filterPodCluster.value) params.pod_cluster = filterPodCluster.value
    rows.value = (await agentApi.list(params)) || []
  } catch {
    rows.value = []
  } finally {
    loading.value = false
  }
}

async function loadUpgrade() {
  loading.value = true
  try {
    upgradeRows.value = (await agentApi.listUpgrade()) || []
  } catch {
    upgradeRows.value = []
  } finally {
    loading.value = false
  }
}

function onViewChange() {
  if (view.value === 'upgrade') loadUpgrade()
  else load()
}

async function ensureGroups() {
  if (groups.value.length) return
  try {
    groups.value = (await agentGroupApi.list()) || []
  } catch {
    groups.value = []
  }
}

async function openEdit(row: Agent) {
  await ensureGroups()
  editName.value = row.NAME
  editForm.enable = true
  editForm.agent_group_id = ''
  editForm.tap_mode = ''
  editVisible.value = true
}

async function submitEdit() {
  const payload: AgentUpdatePayload = { enable: editForm.enable }
  if (editForm.agent_group_id) payload.agent_group_id = editForm.agent_group_id
  if (editForm.tap_mode) payload.tap_mode = editForm.tap_mode
  saving.value = true
  try {
    await agentApi.update(editName.value, payload)
    ElMessage.success('保存成功')
    editVisible.value = false
    await load()
  } finally {
    saving.value = false
  }
}

async function remove(row: Agent) {
  try {
    await ElMessageBox.confirm(`确认删除 Agent "${row.NAME}"？`, '删除确认', { type: 'warning' })
  } catch {
    return
  }
  await agentApi.remove(row.NAME)
  ElMessage.success('删除成功')
  await load()
}

async function rebalance(type: string) {
  // 先 check=false 探测是否需要迁移
  let total = 0
  try {
    const checkRes: any = await agentApi.rebalance(type, false)
    total = checkRes?.TOTAL_SWITCH_VTAP_NUM ?? 0
  } catch {
    return
  }
  if (total === 0) {
    ElMessage.info(`${type} 无需重平衡`)
    return
  }
  try {
    await ElMessageBox.confirm(
      `${type} 重平衡：预计需要迁移 ${total} 个 Agent，确认执行？`,
      '重平衡确认',
      { type: 'warning' },
    )
  } catch {
    return
  }
  const res: any = await agentApi.rebalance(type, true)
  const details = res?.DETAILS
  ElMessage.success(`${type} 重平衡完成` + (Array.isArray(details) ? `，${details.length} 条变更` : ''))
}

onMounted(() => {
  load()
  ensureGroups()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  align-items: center;
  flex-wrap: wrap;
}
.ex-tag {
  margin-right: 4px;
}
</style>
