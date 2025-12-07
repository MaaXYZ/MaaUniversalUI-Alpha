<script setup lang="ts">
  defineProps<{
    checkedCount: number
    totalCount: number
  }>()
</script>

<template>
  <div class="text-gray-800 dark:text-gray-100 font-medium select-none">
    Task List
    <span
      v-if="totalCount > 0"
      class="text-gray-400 dark:text-gray-500 text-sm ml-1"
    >
      ({{ checkedCount }}/{{ totalCount }})
    </span>
  </div>
</template>
