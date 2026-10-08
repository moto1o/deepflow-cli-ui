<template>
  <div class="config-page">
    <div class="left">
      <div class="toolbar">
        <el-button size="small" @click="load">刷新</el-button>
      </div>
      <el-table
        :data="rows"
        v-loading="loading"
        border
        highlight-current-row
        @current-change="onSelect"
        height="calc(100vh - 220px)"
      >
        <el-table-column prop="NAME" label="NAME" />
        <el-table-column prop="AGENT_GROUP_ID" label="ID" />
      </el-table>
    </div>

    <div class="right">
      <!-- 创建区：独立选组 -->
      <el-card class="create-card" shadow="never">
        <template #header><span>创建配置</span></template>
        <div class="create-row">
          <el-select
            v-model="createGroupId"
            filterable
            placeholder="选择目标 Agent 组（已有配置的组不显示）"
            style="width: 320px"
          >
            <el-option
              v-for="g in availableGroups"
              :key="g.SHORT_UUID"
              :label="`${g.NAME} (${g.SHORT_UUID})`"
              :value="g.SHORT_UUID"
            />
          </el-select>
          <el-button size="small" @click="loadTemplate">加载示例模板</el-button>
          <el-button
            size="small"
            type="primary"
            :disabled="!createGroupId"
            :loading="saving"
            @click="save('create')"
          >
            创建
          </el-button>
        </div>
      </el-card>

      <!-- 查看/更新/删除区：从左侧表格选中 -->
      <div class="toolbar">
        <span class="cur">
          已选配置:
          <b>{{ current ? `${current.NAME} (${current.AGENT_GROUP_ID})` : '未选择' }}</b>
        </span>
        <el-button size="small" :disabled="!current" @click="loadYaml">查看配置</el-button>
        <el-button size="small" type="warning" :disabled="!current" :loading="saving" @click="save('update')">
          更新
        </el-button>
        <el-button size="small" type="danger" :disabled="!current" @click="remove">删除配置</el-button>
      </div>
      <YamlEditor v-model="yaml" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import YamlEditor from '../../components/common/YamlEditor.vue'
import { agentGroupConfigApi, type AgentGroupConfig } from '../../api/agentGroupConfig'
import { agentGroupApi, type AgentGroup } from '../../api/agentGroup'

const rows = ref<AgentGroupConfig[]>([])
const loading = ref(false)
const current = ref<AgentGroupConfig | null>(null)
const yaml = ref('')
const saving = ref(false)

// 创建专用
const allGroups = ref<AgentGroup[]>([])
const createGroupId = ref('')

// 过滤掉已有配置的组
const availableGroups = computed(() => {
  const existing = new Set(rows.value.map((r) => r.AGENT_GROUP_ID))
  return allGroups.value.filter((g) => !existing.has(g.SHORT_UUID))
})

async function load() {
  loading.value = true
  try {
    const [configs, groups] = await Promise.all([
      agentGroupConfigApi.list().catch(() => [] as AgentGroupConfig[]),
      agentGroupApi.list().catch(() => [] as AgentGroup[]),
    ])
    rows.value = configs || []
    allGroups.value = groups || []
    // 若当前选中的组已有配置，清空选择
    if (createGroupId.value && !availableGroups.value.find((g) => g.SHORT_UUID === createGroupId.value)) {
      createGroupId.value = ''
    }
  } finally {
    loading.value = false
  }
}

function onSelect(row: AgentGroupConfig | null) {
  current.value = row
}

async function loadTemplate() {
  const res = await agentGroupConfigApi.template()
  yaml.value = res?.yaml || ''
  ElMessage.success('已加载示例模板')
}

async function loadYaml() {
  if (!current.value) return
  const res = await agentGroupConfigApi.getYaml(current.value.AGENT_GROUP_ID)
  yaml.value = res?.yaml || ''
}

async function save(mode: 'create' | 'update') {
  if (!yaml.value.trim()) {
    ElMessage.warning('配置内容为空')
    return
  }
  if (mode === 'create' && !createGroupId.value) {
    ElMessage.warning('请先选择目标 Agent 组')
    return
  }
  if (mode === 'update' && !current.value) return
  const targetId = mode === 'create' ? createGroupId.value : current.value!.AGENT_GROUP_ID
  if (mode === 'update') {
    try {
      await ElMessageBox.confirm(
        `确认更新组 "${current.value!.NAME}" 的配置？此操作将覆盖当前配置。`,
        '更新确认',
        { type: 'warning' },
      )
    } catch {
      return
    }
  }
  saving.value = true
  try {
    if (mode === 'create') {
      await agentGroupConfigApi.create(targetId, yaml.value)
    } else {
      await agentGroupConfigApi.update(targetId, yaml.value)
    }
    ElMessage.success(mode === 'create' ? '创建成功' : '更新成功')
    if (mode === 'create') await load()
  } finally {
    saving.value = false
  }
}

async function remove() {
  if (!current.value) return
  try {
    await ElMessageBox.confirm(`确认删除组 "${current.value.NAME}" 的配置？`, '删除确认', {
      type: 'warning',
    })
  } catch {
    return
  }
  await agentGroupConfigApi.remove(current.value.AGENT_GROUP_ID)
  ElMessage.success('删除成功')
  yaml.value = ''
  current.value = null
  await load()
}

onMounted(load)
</script>

<style scoped>
.config-page {
  display: flex;
  gap: 16px;
  height: 100%;
}
.left {
  width: 320px;
  flex-shrink: 0;
}
.right {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  align-items: center;
  flex-wrap: wrap;
}
.cur {
  margin-right: 8px;
}
.create-card {
  margin-bottom: 12px;
  flex-shrink: 0;
}
.create-card :deep(.el-card__body) {
  padding: 12px 16px;
  overflow: visible;
}
.create-row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: nowrap;
}
</style>
