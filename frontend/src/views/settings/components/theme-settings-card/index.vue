<script setup lang="ts">
  import { computed } from 'vue'
  import { Icon } from '@iconify/vue'
  import { useI18n } from 'vue-i18n'
  import { useConfigStore, type ThemeMode } from '@/store/modules/config'

  const { t } = useI18n()
  const configStore = useConfigStore()

  /** Theme options */
  const themeOptions = computed(() => [
    { value: 'light' as ThemeMode, label: t('settings.theme.light'), icon: 'fluent:weather-sunny-20-regular' },
    { value: 'dark' as ThemeMode, label: t('settings.theme.dark'), icon: 'fluent:weather-moon-20-regular' },
    { value: 'system' as ThemeMode, label: t('settings.theme.system'), icon: 'fluent:desktop-20-regular' },
  ])

  /** Current theme mode */
  const currentMode = computed(() => configStore.theme)

  /** Handle theme selection */
  function handleSelectTheme(mode: ThemeMode) {
    configStore.setTheme(mode)
  }

  /** Check if theme is selected */
  function isSelected(mode: ThemeMode) {
    return currentMode.value === mode
  }
</script>

<template>
  <div class="w-full">
    <!-- Theme Settings Card -->
    <div
      class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-lg"
    >
      <!-- Card Header -->
      <div
        class="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80 rounded-t-xl"
      >
        <div
          class="w-10 h-10 rounded-lg bg-linear-to-br from-amber-500 to-orange-600 flex items-center justify-center shadow-md"
        >
          <Icon
            icon="fluent:color-20-regular"
            width="22"
            height="22"
            class="text-white"
          />
        </div>
        <div class="flex-1">
          <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-100">
            {{ t('settings.theme.title') }}
          </h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('settings.theme.desc') }}
          </p>
        </div>
      </div>

      <!-- Card Content -->
      <div class="p-5">
        <!-- Theme Options -->
        <div class="grid grid-cols-3 gap-3">
          <button
            v-for="option in themeOptions"
            :key="option.value"
            class="flex flex-col items-center gap-2 p-4 rounded-lg border-2 transition-all duration-200 cursor-pointer"
            :class="
              isSelected(option.value)
                ? 'border-amber-500 bg-amber-50 dark:bg-amber-900/20'
                : 'border-gray-200 dark:border-gray-700 hover:border-amber-300 dark:hover:border-amber-600 bg-white dark:bg-gray-700/50'
            "
            @click="handleSelectTheme(option.value)"
          >
            <div
              class="w-10 h-10 rounded-full flex items-center justify-center transition-colors"
              :class="
                isSelected(option.value)
                  ? 'bg-amber-500 text-white'
                  : 'bg-gray-100 dark:bg-gray-600 text-gray-500 dark:text-gray-400'
              "
            >
              <Icon
                :icon="option.icon"
                width="22"
                height="22"
              />
            </div>
            <span
              class="text-sm font-medium transition-colors"
              :class="
                isSelected(option.value)
                  ? 'text-amber-700 dark:text-amber-300'
                  : 'text-gray-600 dark:text-gray-300'
              "
            >
              {{ option.label }}
            </span>
            <!-- Selected indicator -->
            <div
              class="w-4 h-4 rounded-full border-2 flex items-center justify-center transition-colors"
              :class="
                isSelected(option.value)
                  ? 'border-amber-500 bg-amber-500'
                  : 'border-gray-300 dark:border-gray-600'
              "
            >
              <Icon
                v-if="isSelected(option.value)"
                icon="fluent:checkmark-12-regular"
                width="10"
                height="10"
                class="text-white"
              />
            </div>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
