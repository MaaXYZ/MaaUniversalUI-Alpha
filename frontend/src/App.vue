<script setup lang="ts">
  import { onMounted, onUnmounted } from 'vue'
  import Toast from '@/volt/Toast.vue'
  import { usePiStore } from '@/store/modules/pi'
  import { useConfigStore } from '@/store/modules/config'
  import { useTaskListStore } from '@/store/modules/taskList'
  import { useGlobalEvents } from '@/composables'

  const piStore = usePiStore()
  const configStore = useConfigStore()
  const taskListStore = useTaskListStore()

  useGlobalEvents()

  onMounted(async () => {
    await piStore.load()
    await configStore.load()

    taskListStore.loadFromConfig()
    taskListStore.initRunningState()
  })

  onUnmounted(() => {
    taskListStore.cleanupRunningState()
  })
</script>

<template>
  <toast />
  <router-view />
</template>

<style scoped></style>
