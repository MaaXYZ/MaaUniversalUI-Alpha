<script setup lang="ts">
  import { Icon } from '@iconify/vue'
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
    <span class="ml-2">加载中...</span>
  </div>
</template>
