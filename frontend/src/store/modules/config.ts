import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { GetConfig, SaveConfig } from '@wails/go/pi/service'
import {
  GetConfig as GetAppConfig,
  SaveConfig as SaveAppConfig,
  GetSupported,
} from '@wails/go/appconf/service'
import { pi, appconf } from '@wails/go/models'

export type ThemeMode = 'light' | 'dark' | 'system'

export const useConfigStore = defineStore('config', () => {
  // ============ State ============


  const piConfig = ref<pi.InterfaceConfig | null>(null)
  const appConfig = ref<appconf.AppConfig | null>(null)
  const appSupported = ref<appconf.Supported | null>(null)

  // Shared loading/saving state
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)

  /** Whether the system prefers dark mode */
  const systemPrefersDark = ref(false)

  // ============ Getters ============

  /** Whether the config is loaded */
  const isLoaded = computed(
    () => piConfig.value !== null && appConfig.value !== null
  )

  /** Current controller configuration */
  const controller = computed(
    () => piConfig.value?.controller ?? { name: '', type: '' }
  )

  /** Current ADB configuration */
  const adb = computed(() => piConfig.value?.adb ?? null)

  /** Current Win32 configuration */
  const win32 = computed(() => piConfig.value?.win32 ?? null)

  /** Current resource */
  const resource = computed(() => piConfig.value?.resource ?? '')

  /** Task list */
  const tasks = computed(() => piConfig.value?.task ?? [])

  // App config getters
  /** Current theme mode */
  const theme = computed(
    () => (appConfig.value?.theme as ThemeMode) ?? 'system'
  )

  /** Current language */
  const language = computed(() => appConfig.value?.language ?? 'zh-CN')

  /** Whether dark mode is currently active */
  const isDark = computed(() => {
    if (theme.value === 'system') {
      return systemPrefersDark.value
    }
    return theme.value === 'dark'
  })

  /** Current effective theme */
  const effectiveTheme = computed(() => (isDark.value ? 'dark' : 'light'))

  /** Supported themes */
  const supportedThemes = computed(() => appSupported.value?.themes ?? [])

  /** Supported languages */
  const supportedLanguages = computed(() => appSupported.value?.languages ?? [])

  // ============ Actions ============

  /** Initialize system preference detection */
  function initSystemPreference() {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    systemPrefersDark.value = mediaQuery.matches

    mediaQuery.addEventListener('change', (e) => {
      systemPrefersDark.value = e.matches
    })
  }

  /** Apply theme to document */
  function applyTheme() {
    const html = document.documentElement
    if (isDark.value) {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }
  }

  /** Load all configs from backend */
  async function load() {
    if (loading.value) return

    loading.value = true
    error.value = null

    try {
      // Initialize system preference detection
      initSystemPreference()

      // Load all configs in parallel
      const [interfaceConfig, appConfigData, supportedData] = await Promise.all(
        [GetConfig(), GetAppConfig(), GetSupported()]
      )

      piConfig.value = interfaceConfig
      appConfig.value = appConfigData
      appSupported.value = supportedData

      // Apply theme immediately
      applyTheme()
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      console.error('Failed to load config:', e)
    } finally {
      loading.value = false
    }
  }

  /** Save all configs to backend */
  async function save() {
    if (saving.value) return

    saving.value = true
    error.value = null

    try {
      const savePromises: Promise<void>[] = []

      if (piConfig.value) {
        savePromises.push(SaveConfig(piConfig.value))
      }

      if (appConfig.value) {
        savePromises.push(SaveAppConfig(appConfig.value))
      }

      await Promise.all(savePromises)
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      console.error('Failed to save config:', e)
    } finally {
      saving.value = false
    }
  }

  /** Set controller configuration */
  function setController(ctrl: pi.ConfigController) {
    if (piConfig.value) {
      piConfig.value.controller = ctrl
    }
  }

  /** Set ADB configuration */
  function setAdb(adbConfig: pi.ConfigAdb | undefined) {
    if (piConfig.value) {
      piConfig.value.adb = adbConfig
    }
  }

  /** Set Win32 configuration */
  function setWin32(win32Config: pi.ConfigWin32 | undefined) {
    if (piConfig.value) {
      piConfig.value.win32 = win32Config
    }
  }

  /** Set resource */
  function setResource(res: string) {
    if (piConfig.value) {
      piConfig.value.resource = res
    }
  }

  /** Set task list */
  function setTasks(taskList: pi.ConfigTask[]) {
    if (piConfig.value) {
      piConfig.value.task = taskList
    }
  }

  /** Ensure config is initialized */
  function ensureConfig() {
    if (!piConfig.value) {
      piConfig.value = pi.InterfaceConfig.createFrom({
        controller: { name: '', type: '' },
        resource: '',
        task: [],
      })
    }
  }

  /** Set theme mode */
  async function setTheme(newTheme: ThemeMode) {
    if (!appConfig.value) return

    appConfig.value.theme = newTheme
    await save()
  }

  /** Set language */
  async function setLanguage(newLanguage: string) {
    if (!appConfig.value) return

    appConfig.value.language = newLanguage
    await save()
  }

  // Watch for theme changes and apply
  watch(isDark, () => {
    applyTheme()
  })

  return {
    // State
    piConfig,
    appConfig,
    appSupported,
    loading,
    saving,
    error,
    systemPrefersDark,

    // Getters
    isLoaded,
    controller,
    adb,
    win32,
    resource,
    tasks,
    theme,
    language,
    isDark,
    effectiveTheme,
    supportedThemes,
    supportedLanguages,

    // Actions
    load,
    save,
    setController,
    setAdb,
    setWin32,
    setResource,
    setTasks,
    ensureConfig,
    setTheme,
    setLanguage,
    applyTheme,
  }
})
