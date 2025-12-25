<script setup lang="ts">
  import { Icon } from '@iconify/vue'
  import { useI18n } from 'vue-i18n'

  defineProps<{
    hasAvailableTasks: boolean
  }>()

  const emit = defineEmits<{
    addTask: []
  }>()

  const { t } = useI18n()
</script>

<template>
  <div
    class="flex flex-col items-center justify-center h-full text-gray-400 dark:text-gray-500"
  >
    <Icon
      icon="fluent:task-list-20-regular"
      width="48"
      height="48"
    />
    <span class="mt-2 text-sm select-none">{{ t('home.task-list.empty-state.title') }}</span>
    <button
      v-if="hasAvailableTasks"
      class="mt-3 px-4 py-2 text-sm text-indigo-600 dark:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors select-none"
      @click.stop="emit('addTask')"
    >
      {{ t('home.task-list.empty-state.add-task') }}
    </button>
  </div>
</template>
