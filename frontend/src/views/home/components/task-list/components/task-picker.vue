<script setup lang="ts">
  import { ref, onMounted, onUnmounted } from 'vue'
  import { Icon } from '@iconify/vue'
  import { pi } from '@wails/go/models'
  import { usePiStore } from '@/store/modules'

  const props = defineProps<{
    availableTasks: pi.V2Task[]
    hasAvailableTasks: boolean
  }>()

  const emit = defineEmits<{
    addTask: [task: pi.V2Task]
  }>()

  const piStore = usePiStore()

  /** Whether the task picker is visible */
  const showTaskPicker = ref(false)

  /** Task picker container reference */
  const pickerRef = ref<HTMLElement | null>(null)

  /** Click outside to close the picker */
  function handleClickOutside(event: MouseEvent) {
    if (pickerRef.value && !pickerRef.value.contains(event.target as Node)) {
      showTaskPicker.value = false
    }
  }

  onMounted(() => {
    document.addEventListener('click', handleClickOutside)
  })

  onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside)
  })

  /** toggle task picker visibility */
  function togglePicker() {
    if (props.hasAvailableTasks) {
      showTaskPicker.value = !showTaskPicker.value
    }
  }

  /** add task to list */
  function handleAddTask(task: pi.V2Task) {
    emit('addTask', task)
    showTaskPicker.value = false
  }

  /** get task display name */
  function getTaskDisplayName(task: pi.V2Task): string {
    if (task.label) {
      return piStore.resolveI18n(task.label)
    }
    return task.name
  }

  /** Expose togglePicker for parent component */
  defineExpose({
    togglePicker,
  })
</script>

<template>
  <div
    class="relative"
    ref="pickerRef"
  >
    <button
      class="p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-300 transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
      :disabled="!hasAvailableTasks"
      @click="togglePicker"
    >
      <Icon
        icon="fluent:add-20-regular"
        width="20"
        height="20"
      />
    </button>

    <!-- Task Picker Dropdown -->
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 scale-95 -translate-y-1"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 -translate-y-1"
    >
      <div
        v-if="showTaskPicker"
        class="absolute right-0 top-full mt-2 w-64 max-h-80 overflow-y-auto rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xl z-50 scrollbar-thin"
      >
        <div class="p-2 border-b border-gray-200 dark:border-gray-700">
          <span class="text-xs text-gray-500 dark:text-gray-400 select-none"
            >选择要添加的任务</span
          >
        </div>
        <div class="p-1">
          <button
            v-for="task in availableTasks"
            :key="task.name"
            class="w-full px-3 py-2.5 text-left rounded-md hover:bg-indigo-50 dark:hover:bg-indigo-900/20 text-gray-700 dark:text-gray-200 text-sm transition-colors flex items-center gap-2"
            @click="handleAddTask(task)"
          >
            <Icon
              v-if="task.icon"
              :icon="task.icon"
              width="18"
              height="18"
              class="text-gray-500 dark:text-gray-400 shrink-0"
            />
            <Icon
              v-else
              icon="fluent:task-list-square-20-regular"
              width="18"
              height="18"
              class="text-gray-500 dark:text-gray-400 shrink-0"
            />
            <span class="truncate select-none">{{
              getTaskDisplayName(task)
            }}</span>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
  .scrollbar-thin::-webkit-scrollbar {
    width: 6px;
  }

  .scrollbar-thin::-webkit-scrollbar-track {
    background: transparent;
  }

  .scrollbar-thin::-webkit-scrollbar-thumb {
    background: #d1d5db;
    border-radius: 3px;
  }

  .scrollbar-thin::-webkit-scrollbar-thumb:hover {
    background: #9ca3af;
  }

  .dark .scrollbar-thin::-webkit-scrollbar-thumb {
    background: #4b5563;
  }

  .dark .scrollbar-thin::-webkit-scrollbar-thumb:hover {
    background: #6b7280;
  }
</style>
