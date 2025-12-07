<script setup lang="ts">
  import { Icon } from '@iconify/vue'

  defineProps<{
    message: string
  }>()
</script>

<template>
  <div
    class="flex flex-col items-center justify-center h-full text-red-500 dark:text-red-400"
  >
    <Icon
      icon="fluent:warning-20-regular"
      width="32"
      height="32"
    />
    <span class="mt-2 text-sm">{{ message }}</span>
  </div>
</template>
