<script setup lang="ts">
  import { TaskList, OptionsPanel, TaskControlButton } from './components'
</script>

<template>
  <div class="pl-24 flex gap-4 h-full">
    <task-list />
    <options-panel />
    <div class="flex flex-col justify-end">
      <task-control-button />
    </div>
  </div>
</template>

<style scoped></style>
