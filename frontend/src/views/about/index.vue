<script setup lang="ts">
  import { usePiStore } from '@/store'
import { GetMaaVersion } from '@wails/go/engine/service'
import type { system } from '@wails/go/models'
import { GetAppInfo } from '@wails/go/system/service'
  import { onMounted, ref } from 'vue'

  const piStore = usePiStore()
  const maaVersion = ref('')
  const appInfo = ref<system.AppInfo>()

  onMounted(async () => {
    maaVersion.value = await GetMaaVersion()
    appInfo.value = await GetAppInfo()
  })
</script>

<template>
  <div>About</div>
  <div>Project Version: {{ piStore.version  }}</div>
  <div>Maa Version: {{ maaVersion }}</div>
  <div>App Version: {{ appInfo?.version }}</div>
  <div>Build At: {{ appInfo?.build_at }}</div>
  <div>Build OS: {{ appInfo?.build_os }}</div>
  <div>Build Arch: {{ appInfo?.build_arch }}</div>
  <div>Go Version: {{ appInfo?.go_version }}</div>
</template>

<style scoped></style>
