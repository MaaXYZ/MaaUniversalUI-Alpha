<script setup lang="ts">
  import { TaskList, OptionsPanel, TaskControlButton } from './components'
</script>

<template>
  <div class="pl-24 flex gap-4">
    <task-list />
    <options-panel />
    <div class="">
      <TaskControlButton />
    </div>
  </div>
</template>

<style scoped></style>
