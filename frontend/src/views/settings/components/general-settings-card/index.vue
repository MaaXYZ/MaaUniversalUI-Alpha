<script setup lang="ts">
  import { computed } from 'vue'
  import { Icon } from '@iconify/vue'
  import { useConfigStore } from '@/store/modules/config'
  import { useI18n } from 'vue-i18n'
  import Select from 'primevue/select'

  const configStore = useConfigStore()
  const { t } = useI18n()

  const languageOptions = computed(() => configStore.supportedLanguages)

  const currentLanguage = computed({
    get: () => configStore.language,
    set: (val) => configStore.setLanguage(val),
  })
</script>

<template>
  <div class="max-w-md">
    <div
      class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-lg"
    >
      <!-- Header -->
      <div
        class="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80 rounded-t-xl"
      >
        <div
          class="w-10 h-10 rounded-lg bg-linear-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-md"
        >
          <Icon
            icon="fluent:local-language-20-regular"
            width="22"
            height="22"
            class="text-white"
          />
        </div>
        <div class="flex-1">
          <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-100">
            {{ t('settings.general.title') }}
          </h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('settings.general.language.desc') }}
          </p>
        </div>
      </div>

      <!-- Content -->
      <div class="p-5">
        <Select
          v-model="currentLanguage"
          :options="languageOptions"
          optionLabel="name"
          optionValue="code"
          class="w-full"
          :pt="{
            root: 'w-full flex items-center justify-between px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:border-indigo-500 dark:hover:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-colors cursor-pointer',
            overlay: 'bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-xl mt-1',
            list: 'p-1',
            option: ({ context }) => ({
              class: [
                'px-3 py-2 rounded-md cursor-pointer transition-colors flex items-center justify-between',
                context.selected
                  ? 'bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400'
                  : 'text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700',
              ],
            }),
          }"
        >
          <template #option="slotProps">
            <div class="flex items-center justify-between w-full">
              <span>{{ slotProps.option.name }}</span>
              <Icon
                v-if="slotProps.option.code === currentLanguage"
                icon="fluent:checkmark-12-regular"
              />
            </div>
          </template>
        </Select>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
