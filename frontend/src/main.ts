import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import App from './App.vue'
import router from './router'

// Monaco worker（YAML 仅需基础 editor worker）
;(self as any).MonacoEnvironment = {
  getWorker() {
    return new EditorWorker()
  },
}

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
app.mount('#app')
