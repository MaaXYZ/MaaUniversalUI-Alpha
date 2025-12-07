<script setup lang="ts">
  import { Icon } from '@iconify/vue'

  defineProps<{
    hasAvailableTasks: boolean
  }>()

  const emit = defineEmits<{
    addTask: []
  }>()
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
    <span class="mt-2 text-sm select-none">暂无任务</span>
    <button
      v-if="hasAvailableTasks"
      class="mt-3 px-4 py-2 text-sm text-indigo-600 dark:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors select-none"
      @click.stop="emit('addTask')"
    >
      添加任务
    </button>
  </div>
</template>
