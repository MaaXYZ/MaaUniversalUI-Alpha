<script setup lang="ts">
  import { Icon } from '@iconify/vue'

  interface NavItem {
    id: string
    label: string
    icon: string
  }

  defineProps<{
    items: NavItem[]
    activeId: string
  }>()

  const emit = defineEmits<{
    (e: 'select', id: string): void
  }>()
</script>

<template>
  <nav class="flex flex-col gap-1 pr-4">
    <button
      v-for="item in items"
      :key="item.id"
      class="flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all duration-200 text-left group relative"
      :class="[
        activeId === item.id
          ? 'bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400'
          : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200',
      ]"
      @click="emit('select', item.id)"
    >
      <!-- Active Indicator Bar -->
      <div
        v-if="activeId === item.id"
        class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-6 bg-indigo-500 rounded-r-full"
      ></div>

      <Icon
        :icon="item.icon"
        width="20"
        height="20"
        :class="[
          activeId === item.id
            ? 'text-indigo-500'
            : 'text-gray-500 dark:text-gray-500 group-hover:text-gray-700 dark:group-hover:text-gray-300',
        ]"
      />
      <span>{{ item.label }}</span>
    </button>
  </nav>
</template>
