import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { GetConfig, SaveConfig } from '@wails/go/pi/piService'
import { pi } from '@wails/go/models'

export const useConfigStore = defineStore('config', () => {
  // ============ State ============

  const config = ref<pi.InterfaceConfig | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)

  // ============ Getters ============

  /** Whether the config is loaded */
  const isLoaded = computed(() => config.value !== null)

  /** Current controller configuration */
  const controller = computed(
    () => config.value?.controller ?? { name: '', type: '' }
  )

  /** Current ADB configuration */
  const adb = computed(() => config.value?.adb ?? null)

  /** Current Win32 configuration */
  const win32 = computed(() => config.value?.win32 ?? null)

  /** Current resource */
  const resource = computed(() => config.value?.resource ?? '')

  /** Task list */
  const tasks = computed(() => config.value?.task ?? [])

  // ============ Actions ============

  /** Load config from backend */
  async function load() {
    if (loading.value) return

    loading.value = true
    error.value = null

    try {
      config.value = await GetConfig()
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      console.error('Failed to load config:', e)
    } finally {
      loading.value = false
    }
  }

  /** Save config to backend */
  async function save() {
    if (saving.value || !config.value) return

    saving.value = true
    error.value = null

    try {
      await SaveConfig(config.value)
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      console.error('Failed to save config:', e)
    } finally {
      saving.value = false
    }
  }

  /** Set controller configuration */
  function setController(ctrl: pi.ConfigController) {
    if (config.value) {
      config.value.controller = ctrl
    }
  }

  /** Set ADB configuration */
  function setAdb(adbConfig: pi.ConfigAdb | undefined) {
    if (config.value) {
      config.value.adb = adbConfig
    }
  }

  /** Set Win32 configuration */
  function setWin32(win32Config: pi.ConfigWin32 | undefined) {
    if (config.value) {
      config.value.win32 = win32Config
    }
  }

  /** Set resource */
  function setResource(res: string) {
    if (config.value) {
      config.value.resource = res
    }
  }

  /** Set task list */
  function setTasks(taskList: pi.ConfigTask[]) {
    if (config.value) {
      config.value.task = taskList
    }
  }

  /** Ensure config is initialized */
  function ensureConfig() {
    if (!config.value) {
      config.value = pi.InterfaceConfig.createFrom({
        controller: { name: '', type: '' },
        resource: '',
        task: [],
      })
    }
  }

  return {
    // State
    config,
    loading,
    saving,
    error,

    // Getters
    isLoaded,
    controller,
    adb,
    win32,
    resource,
    tasks,

    // Actions
    load,
    save,
    setController,
    setAdb,
    setWin32,
    setResource,
    setTasks,
    ensureConfig,
  }
})
