<script setup lang="ts">
  import { ref } from 'vue'
  import {
    TaskHeader,
    TaskPicker,
    LoadingState,
    ErrorState,
    EmptyState,
    TaskItem,
  } from './components'
  import { usePiStore, useTaskListStore } from '@/store/modules'
  import { pi } from '@wails/go/models'

  const piStore = usePiStore()
  const taskListStore = useTaskListStore()

  /** Task picker component reference */
  const pickerRef = ref<InstanceType<typeof TaskPicker> | null>(null)

  /** set task checked state */
  function handleCheckChange(id: string, checked: boolean) {
    taskListStore.setTaskChecked(id, checked)
  }

  /** remove task from list */
  function handleRemove(id: string) {
    taskListStore.removeTask(id)
  }

  /** select task for configuration */
  function handleSelect(id: string) {
    taskListStore.selectTask(id)
  }

  /** add task to list */
  function handleAddTask(task: pi.V2Task) {
    taskListStore.addTask(task, true)
  }

  /** toggle task picker visibility from empty state */
  function handleEmptyAddTask() {
    pickerRef.value?.togglePicker()
  }

  // ============ Drag and Sort ============

  /** The id of the task being dragged */
  const draggingId = ref<string | null>(null)

  /** The id of the task being dragged over */
  const dragOverId = ref<string | null>(null)

  /** Start dragging */
  function handleDragStart(id: string) {
    draggingId.value = id
  }

  /** End dragging */
  function handleDragEnd() {
    // execute move
    if (
      draggingId.value &&
      dragOverId.value &&
      draggingId.value !== dragOverId.value
    ) {
      const fromIndex = taskListStore.getTaskIndex(draggingId.value)
      const toIndex = taskListStore.getTaskIndex(dragOverId.value)
      taskListStore.moveTask(fromIndex, toIndex)
    }

    // reset state
    draggingId.value = null
    dragOverId.value = null
  }

  /** Drag into target */
  function handleDragEnter(id: string) {
    if (draggingId.value && id !== draggingId.value) {
      dragOverId.value = id
    }
  }
</script>

<template>
  <div
    class="w-87.5 h-150 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col shadow-lg"
  >
    <!-- Header -->
    <div
      class="h-13 flex justify-between items-center rounded-t-xl border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80 px-5 py-4"
    >
      <task-header
        :checked-count="taskListStore.checkedCount"
        :total-count="taskListStore.totalCount"
      />

      <task-picker
        ref="pickerRef"
        :available-tasks="taskListStore.availableTasks"
        :has-available-tasks="taskListStore.hasAvailableTasks"
        @add-task="handleAddTask"
      />
    </div>

    <!-- List Area -->
    <div class="flex-1 overflow-y-auto p-3 space-y-2 scrollbar-thin">
      <!-- Loading State -->
      <loading-state v-if="piStore.loading" />

      <!-- Error State -->
      <error-state
        v-else-if="piStore.error"
        :message="piStore.error"
      />

      <!-- Empty State -->
      <empty-state
        v-else-if="taskListStore.totalCount === 0"
        :has-available-tasks="taskListStore.hasAvailableTasks"
        @add-task="handleEmptyAddTask"
      />

      <!-- Task List -->
      <template v-else>
        <task-item
          v-for="item in taskListStore.taskList"
          :key="item.id"
          :id="item.id"
          :task="item.task"
          :checked="item.checked"
          :dragging="draggingId === item.id"
          :drag-over="dragOverId === item.id"
          :selected="taskListStore.selectedTaskId === item.id"
          @update:checked="handleCheckChange(item.id, $event)"
          @remove="handleRemove"
          @select="handleSelect"
          @dragstart="handleDragStart"
          @dragend="handleDragEnd"
          @dragenter="handleDragEnter"
        />
      </template>
    </div>
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
