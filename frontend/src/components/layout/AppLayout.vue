<template>
  <el-container class="app-layout">
    <el-header height="auto" class="app-header">
      <ConnectionBar />
    </el-header>
    <el-container class="app-body">
      <el-aside width="200px" class="app-aside">
        <AppSidebar />
      </el-aside>
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import ConnectionBar from './ConnectionBar.vue'
import AppSidebar from './AppSidebar.vue'
</script>

<style scoped>
.app-layout {
  height: 100vh;
}
.app-header {
  padding: 0;
}
.app-body {
  height: 100%;
}
.app-aside {
  border-right: 1px solid var(--el-border-color);
}
.app-main {
  padding: 16px;
  overflow: auto;
}
</style>
