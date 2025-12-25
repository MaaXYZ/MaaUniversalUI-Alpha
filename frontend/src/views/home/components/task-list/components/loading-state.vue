<script setup lang="ts">
  import { Icon } from '@iconify/vue'
  import { useI18n } from 'vue-i18n'

  const { t } = useI18n()
</script>

<template>
  <div
    class="flex items-center justify-center h-full text-gray-400 dark:text-gray-500"
  >
    <Icon
      icon="svg-spinners:ring-resize"
      width="24"
      height="24"
    />
    <span class="ml-2">{{ t('home.task-list.loading-state.title') }}</span>
  </div>
</template>
