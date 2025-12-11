<script setup lang="ts">
  import { usePiStore } from '@/store'
import { GetMaaVersion } from '@wails/go/engine/service'
  import { onMounted, ref } from 'vue'

  const piStore = usePiStore()
  const maaVersion = ref('')

  onMounted(async () => {
    maaVersion.value = await GetMaaVersion()
  })
</script>

<template>
  <div>About</div>
  <div>Project Version: {{ piStore.version  }}</div>
  <div>Maa Version: {{ maaVersion }}</div>
</template>

<style scoped></style>
