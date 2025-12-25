<script setup lang="ts">
  import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
  import { Icon } from '@iconify/vue'
  import { usePiStore, useConfigStore } from '@/store/modules'
  import { pi } from '@wails/go/models'

  const piStore = usePiStore()
  const configStore = useConfigStore()

  /** Whether the controller picker is visible */
  const showPicker = ref(false)

  /** Whether advanced settings is expanded */
  const showAdvanced = ref(false)

  /** Picker container reference */
  const pickerRef = ref<HTMLElement | null>(null)

  /** Current selected controller */
  const currentController = computed(() => configStore.controller)

  /** Current selected controller object from pi */
  const currentControllerObj = computed(() =>
    currentController.value?.name
      ? piStore.getControllerByName(currentController.value.name)
      : null
  )

  /** Current controller display name */
  const currentControllerDisplayName = computed(() => {
    if (!currentControllerObj.value) return '请选择控制器'
    return getControllerDisplayName(currentControllerObj.value)
  })

  /** Available controllers list */
  const controllers = computed(() => piStore.controllers)

  /** Whether controllers are available */
  const hasControllers = computed(() => controllers.value.length > 0)

  /** Whether current controller is ADB type */
  const isAdbController = computed(
    () => currentController.value?.type === 'Adb'
  )

  /** Current ADB config */
  const adbConfig = computed(() => configStore.adb)

  /** ADB Path value */
  const adbPath = computed({
    get: () => adbConfig.value?.adb_path ?? '',
    set: (value: string) => updateAdbConfig({ adb_path: value }),
  })

  /** ADB Address value */
  const adbAddress = computed({
    get: () => adbConfig.value?.address ?? '',
    set: (value: string) => updateAdbConfig({ address: value }),
  })

  /** ADB Config JSON string */
  const adbConfigJson = computed({
    get: () => {
      const config = adbConfig.value?.config
      if (!config || Object.keys(config).length === 0) return ''
      return JSON.stringify(config, null, 2)
    },
    set: (value: string) => {
      if (!value.trim()) {
        updateAdbConfig({ config: undefined })
        return
      }
      try {
        const parsed = JSON.parse(value)
        updateAdbConfig({ config: parsed })
      } catch {
        // Invalid JSON, don't update
      }
    },
  })

  /** Whether the config JSON is valid */
  const isConfigJsonValid = computed(() => {
    const value = adbConfigJson.value
    if (!value.trim()) return true
    try {
      JSON.parse(value)
      return true
    } catch {
      return false
    }
  })

  /** Update ADB config */
  function updateAdbConfig(
    partial: Partial<{
      adb_path: string
      address: string
      config: Record<string, unknown> | undefined
    }>
  ) {
    const current = adbConfig.value ?? {}
    configStore.setAdb({
      adb_path: partial.adb_path ?? current.adb_path ?? '',
      address: partial.address ?? current.address ?? '',
      config: partial.config !== undefined ? partial.config : current.config,
    })
  }

  /** Save ADB config with debounce */
  let saveTimeout: ReturnType<typeof setTimeout> | null = null
  function debouncedSave() {
    if (saveTimeout) clearTimeout(saveTimeout)
    saveTimeout = setTimeout(() => {
      configStore.save()
    }, 500)
  }

  /** Get controller display name (use label if available, otherwise use name) */
  function getControllerDisplayName(ctrl: pi.V2Controller) {
    if (ctrl.label) {
      return piStore.resolveI18n(ctrl.label)
    }
    return ctrl.name
  }

  /** Get controller type badge text */
  function getControllerTypeBadge(ctrl: pi.V2Controller) {
    return ctrl.type === 'Adb'
      ? 'ADB'
      : ctrl.type === 'Win32'
      ? 'Win32'
      : ctrl.type
  }

  /** Check if controller is selected */
  function isSelected(ctrl: pi.V2Controller) {
    return ctrl.name === currentController.value?.name
  }

  /** Click outside to close the picker */
  function handleClickOutside(event: MouseEvent) {
    if (pickerRef.value && !pickerRef.value.contains(event.target as Node)) {
      showPicker.value = false
    }
  }

  onMounted(() => {
    document.addEventListener('click', handleClickOutside)
  })

  onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside)
    if (saveTimeout) clearTimeout(saveTimeout)
  })

  /** Toggle picker visibility */
  function togglePicker() {
    if (hasControllers.value) {
      showPicker.value = !showPicker.value
    }
  }

  /** Toggle advanced settings */
  function toggleAdvanced() {
    showAdvanced.value = !showAdvanced.value
  }

  /** Handle controller selection */
  async function handleSelectController(ctrl: pi.V2Controller) {
    configStore.setController({
      name: ctrl.name,
      type: ctrl.type,
    })
    showPicker.value = false
    await configStore.save()
  }

  /** Watch for controller changes and ensure first load has a default */
  watch(
    [() => configStore.isLoaded, () => piStore.isLoaded],
    ([configLoaded, piLoaded]) => {
      const firstController = controllers.value[0]
      if (
        configLoaded &&
        piLoaded &&
        !currentController.value?.name &&
        firstController
      ) {
        configStore.setController({
          name: firstController.name,
          type: firstController.type,
        })
        configStore.save()
      }
    },
    { immediate: true }
  )
</script>

<template>
  <div class="w-full">
    <!-- Controller Settings Card -->
    <div
      class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-lg"
    >
      <!-- Card Header -->
      <div
        class="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80 rounded-t-xl"
      >
        <div
          class="w-10 h-10 rounded-lg bg-linear-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-md"
        >
          <Icon
            icon="fluent:phone-link-setup-20-regular"
            width="22"
            height="22"
            class="text-white"
          />
        </div>
        <div class="flex-1">
          <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-100">
            控制器配置
          </h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            选择连接方式（ADB/Win32）
          </p>
        </div>

        <!-- Saving Indicator -->
        <Icon
          v-if="configStore.saving"
          icon="fluent:spinner-ios-20-regular"
          width="18"
          height="18"
          class="animate-spin text-gray-400"
        />
      </div>

      <!-- Card Content -->
      <div class="p-5">
        <!-- Loading State -->
        <div
          v-if="piStore.loading"
          class="flex items-center justify-center py-6 text-gray-400 dark:text-gray-500"
        >
          <Icon
            icon="fluent:spinner-ios-20-regular"
            width="24"
            height="24"
            class="animate-spin mr-2"
          />
          <span class="text-sm">加载中...</span>
        </div>

        <!-- No Controllers State -->
        <div
          v-else-if="!hasControllers"
          class="flex flex-col items-center justify-center py-6 text-gray-400 dark:text-gray-500"
        >
          <Icon
            icon="fluent:plug-disconnected-20-regular"
            width="40"
            height="40"
            class="mb-2 opacity-60"
          />
          <p class="text-sm">暂无可用控制器</p>
        </div>

        <!-- Controller Picker -->
        <div
          v-else
          ref="pickerRef"
          class="relative"
        >
          <!-- Trigger Button -->
          <button
            class="w-full px-4 py-3 text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-left cursor-pointer transition-all duration-200 hover:border-emerald-400 dark:hover:border-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent flex items-center justify-between gap-2"
            :class="{
              'ring-2 ring-emerald-500 border-transparent': showPicker,
            }"
            @click="togglePicker"
          >
            <div class="flex items-center gap-2 min-w-0">
              <Icon
                v-if="currentControllerObj?.icon"
                :icon="currentControllerObj.icon"
                width="18"
                height="18"
                class="text-emerald-500 shrink-0"
              />
              <Icon
                v-else
                icon="fluent:plug-connected-20-regular"
                width="18"
                height="18"
                class="text-emerald-500 shrink-0"
              />
              <span
                class="truncate"
                :class="
                  currentController?.name
                    ? 'text-gray-700 dark:text-gray-200'
                    : 'text-gray-400 dark:text-gray-500'
                "
              >
                {{ currentControllerDisplayName }}
              </span>
              <!-- Type Badge -->
              <span
                v-if="currentControllerObj"
                class="px-1.5 py-0.5 text-xs font-medium rounded bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300"
              >
                {{ getControllerTypeBadge(currentControllerObj) }}
              </span>
            </div>
            <Icon
              icon="fluent:chevron-down-20-regular"
              width="18"
              height="18"
              class="text-gray-400 shrink-0 transition-transform duration-200"
              :class="{ 'rotate-180': showPicker }"
            />
          </button>

          <!-- Controller Dropdown -->
          <Transition
            enter-active-class="transition duration-150 ease-out"
            enter-from-class="opacity-0 scale-95 -translate-y-1"
            enter-to-class="opacity-100 scale-100 translate-y-0"
            leave-active-class="transition duration-100 ease-in"
            leave-from-class="opacity-100 scale-100 translate-y-0"
            leave-to-class="opacity-0 scale-95 -translate-y-1"
          >
            <div
              v-if="showPicker"
              class="absolute left-0 right-0 top-full mt-2 max-h-64 overflow-y-auto rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xl z-50 scrollbar-thin"
            >
              <div class="p-2 border-b border-gray-200 dark:border-gray-700">
                <span
                  class="text-xs text-gray-500 dark:text-gray-400 select-none"
                >
                  选择控制器类型
                </span>
              </div>
              <div class="p-1">
                <button
                  v-for="ctrl in controllers"
                  :key="ctrl.name"
                  class="w-full px-3 py-2.5 text-left rounded-md text-sm transition-colors flex items-center gap-2"
                  :class="
                    isSelected(ctrl)
                      ? 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300'
                      : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200'
                  "
                  @click="handleSelectController(ctrl)"
                >
                  <Icon
                    v-if="ctrl.icon"
                    :icon="ctrl.icon"
                    width="18"
                    height="18"
                    class="shrink-0"
                    :class="
                      isSelected(ctrl)
                        ? 'text-emerald-500'
                        : 'text-gray-500 dark:text-gray-400'
                    "
                  />
                  <Icon
                    v-else
                    icon="fluent:plug-connected-20-regular"
                    width="18"
                    height="18"
                    class="shrink-0"
                    :class="
                      isSelected(ctrl)
                        ? 'text-emerald-500'
                        : 'text-gray-500 dark:text-gray-400'
                    "
                  />
                  <span class="truncate select-none flex-1">
                    {{ getControllerDisplayName(ctrl) }}
                  </span>
                  <!-- Type Badge -->
                  <span
                    class="px-1.5 py-0.5 text-xs font-medium rounded shrink-0"
                    :class="
                      isSelected(ctrl)
                        ? 'bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300'
                        : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                    "
                  >
                    {{ getControllerTypeBadge(ctrl) }}
                  </span>
                  <Icon
                    v-if="isSelected(ctrl)"
                    icon="fluent:checkmark-20-regular"
                    width="18"
                    height="18"
                    class="text-emerald-500 shrink-0"
                  />
                </button>
              </div>
            </div>
          </Transition>
        </div>

        <!-- ADB Configuration (shown when ADB type is selected) -->
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 -translate-y-2"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 -translate-y-2"
        >
          <div
            v-if="isAdbController"
            class="mt-4 space-y-4"
          >
            <!-- Divider -->
            <div class="border-t border-gray-200 dark:border-gray-700" />

            <!-- ADB Path -->
            <div class="space-y-1.5">
              <label
                class="block text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                ADB 路径
              </label>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                留空则使用系统默认 ADB
              </p>
              <input
                v-model="adbPath"
                type="text"
                placeholder="例如：C:\platform-tools\adb.exe"
                class="w-full px-3 py-2.5 text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all duration-200"
                @input="debouncedSave"
              />
            </div>

            <!-- ADB Address -->
            <div class="space-y-1.5">
              <label
                class="block text-sm font-medium text-gray-700 dark:text-gray-300"
              >
                连接地址
              </label>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                模拟器或设备的 ADB 连接地址
              </p>
              <input
                v-model="adbAddress"
                type="text"
                placeholder="例如：127.0.0.1:5555"
                class="w-full px-3 py-2.5 text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all duration-200"
                @input="debouncedSave"
              />
            </div>

            <!-- Advanced Settings Toggle -->
            <div class="pt-2">
              <button
                class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-emerald-600 dark:hover:text-emerald-400 transition-colors"
                @click="toggleAdvanced"
              >
                <Icon
                  icon="fluent:chevron-right-20-regular"
                  width="16"
                  height="16"
                  class="transition-transform duration-200"
                  :class="{ 'rotate-90': showAdvanced }"
                />
                <span>高级设置</span>
              </button>

              <!-- Advanced Config -->
              <Transition
                enter-active-class="transition duration-200 ease-out"
                enter-from-class="opacity-0 max-h-0"
                enter-to-class="opacity-100 max-h-96"
                leave-active-class="transition duration-150 ease-in"
                leave-from-class="opacity-100 max-h-96"
                leave-to-class="opacity-0 max-h-0"
              >
                <div
                  v-if="showAdvanced"
                  class="mt-3 space-y-1.5"
                >
                  <label
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    额外配置 (JSON)
                  </label>
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    ADB 控制器的额外配置参数
                  </p>
                  <textarea
                    v-model="adbConfigJson"
                    rows="4"
                    placeholder='例如：{ "extras": {} }'
                    class="w-full px-3 py-2.5 text-sm font-mono bg-white dark:bg-gray-700 border rounded-lg text-gray-700 dark:text-gray-200 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all duration-200 resize-none"
                    :class="
                      isConfigJsonValid
                        ? 'border-gray-300 dark:border-gray-600'
                        : 'border-red-500 dark:border-red-500'
                    "
                    @input="debouncedSave"
                  />
                  <p
                    v-if="!isConfigJsonValid"
                    class="text-xs text-red-500"
                  >
                    JSON 格式无效
                  </p>
                </div>
              </Transition>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
