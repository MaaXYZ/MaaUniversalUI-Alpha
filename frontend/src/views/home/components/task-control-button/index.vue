<script setup lang="ts">
import { useTaskListStore } from '@/store/modules/taskList'
import { storeToRefs } from 'pinia'

const taskListStore = useTaskListStore()
const { isRunning } = storeToRefs(taskListStore)

const handleClick = async () => {
  await taskListStore.toggleEngine()
}
</script>

<template>
  <div
    class="w-60 h-13 flex justify-center items-center rounded-2xl cursor-pointer transition-all duration-300 select-none"
    :class="isRunning
      ? 'bg-rose-500 hover:bg-rose-600'
      : 'bg-emerald-500 hover:bg-emerald-600'"
    @click="handleClick"
  >
    <span class="text-white text-lg font-bold">
      {{ isRunning ? '停止' : '启动' }}
    </span>
  </div>
</template>

<style scoped>

</style>