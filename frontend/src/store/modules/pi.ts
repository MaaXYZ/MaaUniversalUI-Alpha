import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { V2Loaded as LoadV2, GetVersion } from '@wails/go/pi/piService'
import { pi } from '@wails/go/models'

export const PI_VERSION = {
  UNKNOWN: 0,
  V1: 1,
  V2: 2,
} as const

export const usePiStore = defineStore('pi', () => {
  // ============ State ============
  
  const loaded = ref<pi.V2Loaded | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const interfaceVersion = ref<number>(PI_VERSION.UNKNOWN)

  // ============ Getters ============

  /** Whether the PI is loaded */
  const isLoaded = computed(() => loaded.value !== null)

  /** Project interface information */
  const piInterface = computed(() => loaded.value?.Interface ?? null)

  /** Base path */
  const basePath = computed(() => loaded.value?.BasePath ?? '')

  /** i18n resolver */
  const resolvers = computed(() => loaded.value?.Resolvers ?? {})

  /** Task list */
  const tasks = computed(() => piInterface.value?.task ?? [])

  /** Option configuration */
  const options = computed(() => piInterface.value?.option ?? {})

  /** Controller list */
  const controllers = computed(() => piInterface.value?.controller ?? [])

  /** Resource list */
  const resources = computed(() => piInterface.value?.resource ?? [])

  /** Agent configuration */
  const agent = computed(() => piInterface.value?.agent ?? null)

  /** Project name */
  const name = computed(() => piInterface.value?.name ?? '')

  /** Project label */
  const label = computed(
    () => piInterface.value?.label ?? piInterface.value?.name ?? ''
  )

  /** Project title */
  const title = computed(() => piInterface.value?.title ?? '')

  /** Project version */
  const version = computed(() => piInterface.value?.version ?? '')

  /** Project description */
  const description = computed(() => piInterface.value?.description ?? '')

  /** Welcome message */
  const welcome = computed(() => piInterface.value?.welcome ?? '')

  /** Project icon */
  const icon = computed(() => piInterface.value?.icon ?? '')

  /** GitHub address */
  const github = computed(() => piInterface.value?.github ?? '')

  /** Contact information */
  const contact = computed(() => piInterface.value?.contact ?? '')

  /** License */
  const license = computed(() => piInterface.value?.license ?? '')

  /** MirrorChyan RID */
  const mirrorchyanRid = computed(
    () => piInterface.value?.mirrorchyan_rid ?? ''
  )

  /** Protocol version */
  const isV2 = computed(() => interfaceVersion.value === PI_VERSION.V2)

  // ============ Actions ============

  /** Load PI data */
  async function load() {
    if (loading.value) return

    loading.value = true
    error.value = null

    try {
      // get version first
      const ver = await GetVersion()
      interfaceVersion.value = ver

      // only V2 version can get V2Loaded
      if (ver === PI_VERSION.V2) {
        loaded.value = await LoadV2()
      } else if (ver === PI_VERSION.V1) {
        error.value = 'v1 project interface is not supported'
        console.warn('v1 project interface is not supported')
      } else {
        error.value = 'unknown project interface version'
        console.error('Unknown PI version:', ver)
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      console.error('Failed to load PI:', e)
    } finally {
      loading.value = false
    }
  }

  /** Get task by name */
  function getTaskByName(name: string): pi.V2Task | null {
    return tasks.value.find((t: pi.V2Task) => t.name === name) ?? null
  }

  /** Get option configuration by name */
  function getOptionByName(name: string): pi.V2Option | null {
    return options.value[name] ?? null
  }

  /** Get controller by name */
  function getControllerByName(name: string): pi.V2Controller | null {
    return (
      controllers.value.find((c: pi.V2Controller) => c.name === name) ?? null
    )
  }

  /** Get resource by name */
  function getResourceByName(name: string): pi.V2Resource | null {
    return resources.value.find((r: pi.V2Resource) => r.name === name) ?? null
  }

  /** Get task options configuration list */
  function getTaskOptions(task: pi.V2Task): pi.V2Option[] {
    if (!task.option) return []
    return task.option
      .map((optName: string) => getOptionByName(optName))
      .filter(Boolean) as pi.V2Option[]
  }

  /** Get default checked tasks list */
  function getDefaultCheckedTasks(): pi.V2Task[] {
    return tasks.value.filter((t: pi.V2Task) => t.default_check)
  }

  /** Resolve i18n string */
  function resolveI18n(str: string | undefined, locale = 'zh-CN'): string {
    if (!str) return ''
    // if not i18n string (not start with $), return directly
    if (!str.startsWith('$') || str.length === 1) return str

    const key = str.slice(1)
    const resolver = resolvers.value[locale]
    if (resolver && typeof resolver === 'object' && key in resolver) {
      return (resolver as Record<string, string>)[key] ?? str
    }

    // try to use languages in interface
    const languages = piInterface.value?.languages
    if (languages && key in languages) {
      return languages[key] ?? str
    }

    return str
  }

  /** Reset store */
  function reset() {
    loaded.value = null
    loading.value = false
    error.value = null
    interfaceVersion.value = PI_VERSION.UNKNOWN
  }

  return {
    // State
    loaded,
    loading,
    error,
    interfaceVersion,

    // Getters
    isLoaded,
    isV2,
    piInterface,
    basePath,
    resolvers,
    tasks,
    options,
    controllers,
    resources,
    agent,
    name,
    label,
    title,
    version,
    description,
    welcome,
    icon,
    github,
    contact,
    license,
    mirrorchyanRid,

    // Actions
    load,
    getTaskByName,
    getOptionByName,
    getControllerByName,
    getResourceByName,
    getTaskOptions,
    getDefaultCheckedTasks,
    resolveI18n,
    reset,
  }
})
