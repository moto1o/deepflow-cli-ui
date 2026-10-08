<template>
  <el-menu :default-active="activeIndex" router class="sidebar-menu">
    <el-menu-item index="/agents">
      <el-icon><Monitor /></el-icon>
      <span>Agent</span>
    </el-menu-item>
    <el-menu-item index="/agent-groups">
      <el-icon><Collection /></el-icon>
      <span>Agent 组</span>
    </el-menu-item>
    <el-menu-item index="/agent-group-configs">
      <el-icon><Document /></el-icon>
      <span>Agent 组配置</span>
    </el-menu-item>
  </el-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const activeIndex = computed(() => route.path)
</script>

<style scoped>
.sidebar-menu {
  height: 100%;
  border-right: none;
}
</style>
