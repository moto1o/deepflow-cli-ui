<template>
  <div ref="container" class="yaml-editor"></div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import 'monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution'

const props = defineProps<{ modelValue: string; readonly?: boolean }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const container = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null

onMounted(() => {
  if (!container.value) return
  editor = monaco.editor.create(container.value, {
    value: props.modelValue,
    language: 'yaml',
    theme: 'vs',
    automaticLayout: true,
    minimap: { enabled: false },
    readOnly: props.readonly ?? false,
    scrollBeyondLastLine: false,
    tabSize: 2,
  })
  editor.onDidChangeModelContent(() => {
    const v = editor!.getValue()
    if (v !== props.modelValue) emit('update:modelValue', v)
  })
})

watch(
  () => props.modelValue,
  (v) => {
    if (editor && v !== editor.getValue()) editor.setValue(v ?? '')
  },
)

watch(
  () => props.readonly,
  (ro) => {
    editor?.updateOptions({ readOnly: ro ?? false })
  },
)

onBeforeUnmount(() => {
  editor?.dispose()
  editor = null
})
</script>

<style scoped>
.yaml-editor {
  width: 100%;
  height: 100%;
  min-height: 400px;
  border: 1px solid var(--el-border-color);
}
</style>
