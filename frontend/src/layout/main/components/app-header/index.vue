<script setup lang="ts">
  import HeaderLogo from '@/layout/main/components/header-logo/index.vue'
  import HeaderWindowControls from '@/layout/main/components/header-window-controls/index.vue'
</script>

<template>
  <div
    class="flex justify-between px-7.5 h-15"
    style="--wails-draggable: drag"
  >
    <header-logo />
    <header-window-controls />
  </div>
</template>

<style scoped></style>
