<script setup lang="ts">
  import { Icon } from '@iconify/vue'
  import { onMounted, ref } from 'vue'
  import {
    WindowSetAlwaysOnTop,
    WindowMaximise,
    WindowUnmaximise,
    WindowIsMaximised,
    WindowMinimise,
    Quit,
  } from '@wails/runtime'

  const isPinned = ref(false)
  const isMaximized = ref(false)

  const checkIsMaximized = async () => {
    isMaximized.value = await WindowIsMaximised()
  }

  const togglePin = () => {
    WindowSetAlwaysOnTop(!isPinned.value)
    isPinned.value = !isPinned.value
  }

  const minimize = () => {
    WindowMinimise()
  }

  const toggleMaximize = () => {
    if (isMaximized.value) {
      WindowUnmaximise()
    } else {
      WindowMaximise()
    }
    isMaximized.value = !isMaximized.value
  }

  const close = () => {
    Quit()
  }

  onMounted(() => {
    checkIsMaximized()
  })
</script>

<template>
  <div class="flex gap-x-5 my-5" style="--wails-draggable: no-drag">
    <div
      class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors duration-200"
      @click="togglePin"
      :title="isPinned ? '取消置顶' : '置顶'"
    >
      <Icon
        v-if="isPinned"
        icon="fluent:pin-off-20-regular"
        width="20"
        height="20"
      />
      <Icon v-else icon="fluent:pin-20-regular" width="20" height="20" />
    </div>

    <div class="border-l border-gray-500 dark:border-gray-400"></div>

    <div
      class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors duration-200"
      @click="minimize"
      title="最小化"
    >
      <Icon icon="fluent:subtract-20-regular" width="20" height="20" />
    </div>

    <div
      class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors duration-200"
      @click="toggleMaximize"
      :title="isMaximized ? '恢复' : '最大化'"
    >
      <Icon
        v-if="isMaximized"
        icon="fluent:restore-16-regular"
        width="20"
        height="20"
      />
      <Icon v-else icon="fluent:square-20-regular" width="20" height="20" />
    </div>

    <div
      class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors duration-200"
      @click="close"
      title="关闭"
    >
      <Icon icon="fluent:dismiss-20-regular" width="20" height="20" />
    </div>
  </div>
</template>

<style scoped></style>
