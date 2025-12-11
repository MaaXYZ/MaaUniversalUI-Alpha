import { onMounted, onUnmounted } from 'vue'
import { EventsOn } from '@wails/runtime/runtime'
import { useToast } from 'primevue'

interface ToastConfig {
  summary: string
  life: number
}

const defaultConfig: Record<string, ToastConfig> = {
  error: { summary: '错误', life: 5000 },
  warn: { summary: '警告', life: 4000 },
  success: { summary: '成功', life: 3000 },
  info: { summary: '提示', life: 3000 },
}

/**
 * 全局事件监听 composable
 * 自动订阅 app:error, app:warn, app:success, app:info 事件
 * 并在组件卸载时自动清理
 */
export function useGlobalEvents() {
  const toast = useToast()
  const cleanups: Array<() => void> = []

  /** 注册事件监听 */
  function registerEvent(
    eventName: string,
    severity: 'error' | 'warn' | 'success' | 'info'
  ) {
    const config = defaultConfig[severity]
    if (!config) return

    const cleanup = EventsOn(eventName, (message: string) => {
      toast.add({
        severity,
        summary: config.summary,
        detail: message,
        life: config.life,
      })
    })
    cleanups.push(cleanup)
  }

  /** 初始化所有全局事件 */
  function init() {
    registerEvent('app:error', 'error')
    registerEvent('app:warn', 'warn')
    registerEvent('app:success', 'success')
    registerEvent('app:info', 'info')
  }

  /** 清理所有事件监听 */
  function cleanup() {
    cleanups.forEach((fn) => fn())
    cleanups.length = 0
  }

  // 自动在生命周期中处理
  onMounted(() => init())
  onUnmounted(() => cleanup())

  return {
    init,
    cleanup,
  }
}

