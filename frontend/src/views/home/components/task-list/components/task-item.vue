<script setup lang="ts">
  import { computed } from 'vue'
  import { Icon } from '@iconify/vue'
  import { pi } from '@wails/go/models'
  import { usePiStore } from '@/store/modules/pi'

  const props = defineProps<{
    id: string
    task: pi.V2Task
    checked?: boolean
    dragging?: boolean
    dragOver?: boolean
    selected?: boolean
  }>()

  const emit = defineEmits<{
    (e: 'update:checked', value: boolean): void
    (e: 'remove', id: string): void
    (e: 'dragstart', id: string): void
    (e: 'dragend'): void
    (e: 'dragenter', id: string): void
    (e: 'select', id: string): void
  }>()

  const piStore = usePiStore()

  /** get display name: use label first, then name */
  const displayName = computed(() => {
    const label = props.task.label
    if (label) {
      return piStore.resolveI18n(label)
    }
    return props.task.name
  })

  function handleCheckChange(event: Event) {
    const target = event.target as HTMLInputElement
    emit('update:checked', target.checked)
  }

  function handleRemove() {
    emit('remove', props.id)
  }

  function handleSelect() {
    emit('select', props.id)
  }
</script>

<template>
  <div
    class="group w-full h-14 rounded-lg flex justify-between items-center px-4 bg-gray-50 dark:bg-gray-700/50 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 border border-transparent hover:border-indigo-200 dark:hover:border-indigo-800 transition-all duration-200 cursor-grab active:cursor-grabbing"
    :class="{
      'opacity-50 scale-95': dragging,
      'border-indigo-400 dark:border-indigo-500 bg-indigo-50 dark:bg-indigo-900/30':
        dragOver,
      'border-indigo-500! dark:border-indigo-400! bg-indigo-100! dark:bg-indigo-900/40! ring-2 ring-indigo-500/20':
        selected,
    }"
    draggable="true"
    @click="handleSelect"
    @dragstart="emit('dragstart', id)"
    @dragend="emit('dragend')"
    @dragenter.prevent="emit('dragenter', id)"
    @dragover.prevent
  >
    <div class="flex items-center gap-x-3">
      <!-- drag handle -->
      <Icon
        icon="fluent:re-order-dots-vertical-20-regular"
        width="16"
        height="16"
        class="text-gray-400 dark:text-gray-500 cursor-grab"
      />
      <input
        type="checkbox"
        :checked="checked"
        class="size-4.5 bg-white dark:bg-gray-600 border-2 border-gray-300 dark:border-gray-500 rounded focus:ring-2 cursor-pointer hover:scale-110 text-indigo-600 focus:ring-indigo-500 focus:ring-offset-0 transition-transform duration-150"
        @change="handleCheckChange"
      />
      <span
        class="text-gray-700 dark:text-gray-200 text-sm font-medium select-none"
      >
        {{ displayName }}
      </span>
    </div>
    <button
      class="p-1.5 rounded-lg opacity-0 group-hover:opacity-100 hover:bg-red-100 dark:hover:bg-red-900/30 text-gray-400 hover:text-red-500 dark:hover:text-red-400 transition-all duration-200 cursor-pointer"
      @click.stop="handleRemove"
    >
      <Icon
        icon="fluent:delete-20-regular"
        width="18"
        height="18"
      />
    </button>
  </div>
</template>

<style scoped></style>
