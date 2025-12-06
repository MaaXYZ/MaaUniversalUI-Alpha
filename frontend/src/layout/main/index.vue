<script setup lang="ts">
  import AppHeader from '@/layout/main/components/app-header/index.vue'
  import AppSider from '@/layout/main/components/app-sider/index.vue'
</script>

<template>
  <div class="flex flex-col h-screen">
    <app-header />
    <div class="flex-1 relative">
      <div class="fixed top-1/2 -translate-y-1/2 left-7.5 z-10">
        <app-sider />
      </div>
      <div class=" px-7.5">
      <router-view />
      </div>
    </div>
  </div>
</template>

<style scoped></style>
