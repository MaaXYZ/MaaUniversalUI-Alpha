<script setup lang="ts">
  import { useTaskListStore } from '@/store/modules/taskList'
  import { storeToRefs } from 'pinia'
  import { ref, watch } from 'vue'

  const taskListStore = useTaskListStore()
  const { isRunning } = storeToRefs(taskListStore)
  const isStopping = ref<boolean>(false)

  watch(isRunning, (newVal, oldVal) => {
    if (oldVal === true && newVal === false) {
      isStopping.value = false
    }
  })

  const handleClick = async () => {
    if (isStopping.value) return

    if (isRunning.value) {
      isStopping.value = true
    }

    await taskListStore.toggleEngine()
  }
</script>

<template>
  <div
    class="w-60 h-13 flex justify-center items-center rounded-2xl transition-all duration-300 select-none"
    :class="[
      isStopping
        ? 'bg-gray-400 cursor-not-allowed'
        : isRunning
        ? 'bg-rose-500 hover:bg-rose-600 cursor-pointer'
        : 'bg-emerald-500 hover:bg-emerald-600 cursor-pointer',
    ]"
    @click="handleClick"
  >
    <span class="text-white text-lg font-bold">
      {{ isStopping ? '停止中...' : isRunning ? '停止' : '启动' }}
    </span>
  </div>
</template>

<style scoped></style>
