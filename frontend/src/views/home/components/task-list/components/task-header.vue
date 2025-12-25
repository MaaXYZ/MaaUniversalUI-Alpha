<script setup lang="ts">
  import { useI18n } from 'vue-i18n'

  defineProps<{
    checkedCount: number
    totalCount: number
  }>()

  const { t } = useI18n()
</script>

<template>
  <div class="text-gray-800 dark:text-gray-100 font-medium select-none">
    {{ t('home.task-list.title') }}
    <span
      v-if="totalCount > 0"
      class="text-gray-400 dark:text-gray-500 text-sm ml-1"
    >
      ({{ checkedCount }}/{{ totalCount }})
    </span>
  </div>
</template>
