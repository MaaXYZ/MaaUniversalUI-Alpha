<script setup lang="ts">
  import {
    ResSettingsCard,
    CtrlSettingsCard,
    ThemeSettingsCard,
    GeneralSettingsCard,
  } from './components'
</script>

<template>
  <div class="pl-24 space-y-4">
    <general-settings-card />
    <theme-settings-card />
    <res-settings-card />
    <ctrl-settings-card />
  </div>
</template>
