<script setup lang="ts">
  import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
  import { Icon } from '@iconify/vue'
  import { useI18n } from 'vue-i18n'
  import { usePiStore, useConfigStore } from '@/store/modules'
  import { pi } from '@wails/go/models'

  const { t } = useI18n()
  const piStore = usePiStore()
  const configStore = useConfigStore()

  /** Whether the resource picker is visible */
  const showPicker = ref(false)

  /** Picker container reference */
  const pickerRef = ref<HTMLElement | null>(null)

  /** Current selected resource name */
  const currentResource = computed(() => configStore.resource)

  /** Current selected resource object */
  const currentResourceObj = computed(() =>
    currentResource.value
      ? piStore.getResourceByName(currentResource.value)
      : null
  )

  /** Current resource display name */
  const currentResourceDisplayName = computed(() => {
    if (!currentResourceObj.value) return t('settings.resource.select_resource')
    return getResourceDisplayName(currentResourceObj.value)
  })

  /** Available resources list */
  const resources = computed(() => piStore.resources)

  /** Whether resources are available */
  const hasResources = computed(() => resources.value.length > 0)

  /** Get resource display name (use label if available, otherwise use name) */
  function getResourceDisplayName(res: pi.V2Resource) {
    if (res.label) {
      return piStore.resolveI18n(res.label)
    }
    return res.name
  }

  /** Check if resource is selected */
  function isSelected(res: pi.V2Resource) {
    return res.name === currentResource.value
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
  })

  /** Toggle picker visibility */
  function togglePicker() {
    if (hasResources.value) {
      showPicker.value = !showPicker.value
    }
  }

  /** Handle resource selection */
  async function handleSelectResource(res: pi.V2Resource) {
    configStore.setResource(res.name)
    showPicker.value = false
    await configStore.save()
  }

  /** Watch for resource changes and ensure first load has a default */
  watch(
    [() => configStore.isLoaded, () => piStore.isLoaded],
    ([configLoaded, piLoaded]) => {
      const firstResource = resources.value[0]
      if (configLoaded && piLoaded && !currentResource.value && firstResource) {
        configStore.setResource(firstResource.name)
        configStore.save()
      }
    },
    { immediate: true }
  )
</script>

<template>
  <div class="w-full">
    <!-- Resource Settings Card -->
    <div
      class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-lg"
    >
      <!-- Card Header -->
      <div
        class="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80 rounded-t-xl"
      >
        <div
          class="w-10 h-10 rounded-lg bg-linear-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-md"
        >
          <Icon
            icon="fluent:folder-open-20-regular"
            width="22"
            height="22"
            class="text-white"
          />
        </div>
        <div class="flex-1">
          <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-100">
            {{ t('settings.resource.title') }}
          </h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('settings.resource.desc') }}
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
          <span class="text-sm">{{ t('common.loading') }}</span>
        </div>

        <!-- No Resources State -->
        <div
          v-else-if="!hasResources"
          class="flex flex-col items-center justify-center py-6 text-gray-400 dark:text-gray-500"
        >
          <Icon
            icon="fluent:folder-prohibited-20-regular"
            width="40"
            height="40"
            class="mb-2 opacity-60"
          />
          <p class="text-sm">{{ t('settings.resource.no_resources') }}</p>
        </div>

        <!-- Resource Picker -->
        <div
          v-else
          ref="pickerRef"
          class="relative"
        >
          <!-- Trigger Button -->
          <button
            class="w-full px-4 py-3 text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-left cursor-pointer transition-all duration-200 hover:border-indigo-400 dark:hover:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent flex items-center justify-between gap-2"
            :class="{ 'ring-2 ring-indigo-500 border-transparent': showPicker }"
            @click="togglePicker"
          >
            <div class="flex items-center gap-2 min-w-0">
              <Icon
                v-if="currentResourceObj?.icon"
                :icon="currentResourceObj.icon"
                width="18"
                height="18"
                class="text-indigo-500 shrink-0"
              />
              <Icon
                v-else
                icon="fluent:folder-20-regular"
                width="18"
                height="18"
                class="text-indigo-500 shrink-0"
              />
              <span
                class="truncate"
                :class="
                  currentResource
                    ? 'text-gray-700 dark:text-gray-200'
                    : 'text-gray-400 dark:text-gray-500'
                "
              >
                {{ currentResourceDisplayName }}
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

          <!-- Resource Dropdown -->
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
                  {{ t('settings.resource.select_version') }}
                </span>
              </div>
              <div class="p-1">
                <button
                  v-for="res in resources"
                  :key="res.name"
                  class="w-full px-3 py-2.5 text-left rounded-md text-sm transition-colors flex items-center gap-2"
                  :class="
                    isSelected(res)
                      ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300'
                      : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200'
                  "
                  @click="handleSelectResource(res)"
                >
                  <Icon
                    v-if="res.icon"
                    :icon="res.icon"
                    width="18"
                    height="18"
                    class="shrink-0"
                    :class="
                      isSelected(res)
                        ? 'text-indigo-500'
                        : 'text-gray-500 dark:text-gray-400'
                    "
                  />
                  <Icon
                    v-else
                    icon="fluent:folder-20-regular"
                    width="18"
                    height="18"
                    class="shrink-0"
                    :class="
                      isSelected(res)
                        ? 'text-indigo-500'
                        : 'text-gray-500 dark:text-gray-400'
                    "
                  />
                  <span class="truncate select-none flex-1">
                    {{ getResourceDisplayName(res) }}
                  </span>
                  <Icon
                    v-if="isSelected(res)"
                    icon="fluent:checkmark-20-regular"
                    width="18"
                    height="18"
                    class="text-indigo-500 shrink-0"
                  />
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
