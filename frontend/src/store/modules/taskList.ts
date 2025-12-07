import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { usePiStore } from './pi'
import { useConfigStore } from './config'
import { pi } from '@wails/go/models'

/** Task item (contains task information and checked state) */
export interface TaskListItem {
  id: string
  task: pi.V2Task
  checked: boolean
}

/** Generate UUID */
function generateUUID(): string {
  return crypto.randomUUID()
}

export const useTaskListStore = defineStore('taskList', () => {
  const piStore = usePiStore()
  const configStore = useConfigStore()

  // ============ State ============

  /** Task list (tasks added by user) */
  const taskList = ref<TaskListItem[]>([])

  // ============ Getters ============

  /** Available tasks list (all tasks in pi can be added repeatedly) */
  const availableTasks = computed(() => piStore.tasks)

  /** Checked tasks list */
  const checkedTasks = computed(() => {
    return taskList.value
      .filter((item) => item.checked)
      .map((item) => item.task)
  })

  /** Checked tasks list (contains id) */
  const checkedItems = computed(() => {
    return taskList.value.filter((item) => item.checked)
  })

  /** Checked tasks count */
  const checkedCount = computed(
    () => taskList.value.filter((item) => item.checked).length
  )

  /** Total tasks count */
  const totalCount = computed(() => taskList.value.length)

  /** Whether there are still tasks that can be added */
  const hasAvailableTasks = computed(() => piStore.tasks.length > 0)

  // ============ Config Sync ============

  /** Load task list from config */
  function loadFromConfig() {
    const configTasks = configStore.tasks
    const piTasks = piStore.tasks

    // clear existing task list
    taskList.value = []

    // restore task list from config
    for (const configTask of configTasks) {
      const piTask = piTasks.find((t: pi.V2Task) => t.name === configTask.name)
      if (piTask) {
        taskList.value.push({
          id: configTask.id,
          task: piTask,
          checked: configTask.checked,
        })
      }
    }
  }

  /** Sync task list to config and save */
  async function syncToConfig() {
    // update task list in config
    configStore.setTasks(
      taskList.value.map((item) =>
        pi.ConfigTask.createFrom({
          id: item.id,
          name: item.task.name,
          checked: item.checked,
          option: [],
        })
      )
    )

    // save to backend
    await configStore.save()
  }

  // ============ Actions ============

  /** Initialize task list (add default tasks according to default_check) */
  function initFromPi() {
    const defaultTasks = piStore.getDefaultCheckedTasks()
    taskList.value = defaultTasks.map((task) => ({
      id: generateUUID(),
      task,
      checked: true,
    }))
  }

  /** Add task to list (same task can be added repeatedly) */
  function addTask(task: pi.V2Task, checked = false): string {
    const id = generateUUID()
    taskList.value.push({ id, task, checked })
    syncToConfig()
    return id
  }

  /** Add tasks in batch */
  function addTasks(tasks: pi.V2Task[], checked = false): string[] {
    const ids = tasks.map((task) => {
      const id = generateUUID()
      taskList.value.push({ id, task, checked })
      return id
    })
    syncToConfig()
    return ids
  }

  /** Remove task from list by id */
  function removeTask(id: string) {
    const index = taskList.value.findIndex((item) => item.id === id)
    if (index !== -1) {
      taskList.value.splice(index, 1)
      syncToConfig()
      return true
    }
    return false
  }

  /** Set task checked state by id */
  function setTaskChecked(id: string, checked: boolean) {
    const item = taskList.value.find((item) => item.id === id)
    if (item) {
      item.checked = checked
      syncToConfig()
    }
  }

  /** Toggle task checked state by id */
  function toggleTaskChecked(id: string) {
    const item = taskList.value.find((item) => item.id === id)
    if (item) {
      item.checked = !item.checked
      syncToConfig()
    }
  }

  /** Check all */
  function checkAll() {
    taskList.value.forEach((item) => {
      item.checked = true
    })
    syncToConfig()
  }

  /** Uncheck all */
  function uncheckAll() {
    taskList.value.forEach((item) => {
      item.checked = false
    })
    syncToConfig()
  }

  /** Clear task list */
  function clear() {
    taskList.value = []
  }

  /** Reset (clear and re-initialize from PI) */
  function reset() {
    clear()
    initFromPi()
    syncToConfig()
  }

  /** Move task position (drag and sort) */
  function moveTask(fromIndex: number, toIndex: number) {
    if (fromIndex === toIndex) return
    if (fromIndex < 0 || fromIndex >= taskList.value.length) return
    if (toIndex < 0 || toIndex >= taskList.value.length) return

    const item = taskList.value.splice(fromIndex, 1)[0]
    if (item) {
      taskList.value.splice(toIndex, 0, item)
      syncToConfig()
    }
  }

  /** Get task index by id */
  function getTaskIndex(id: string): number {
    return taskList.value.findIndex((item) => item.id === id)
  }

  return {
    // State
    taskList,

    // Getters
    availableTasks,
    checkedTasks,
    checkedItems,
    checkedCount,
    totalCount,
    hasAvailableTasks,

    // Config Sync
    loadFromConfig,
    syncToConfig,

    // Actions
    initFromPi,
    addTask,
    addTasks,
    removeTask,
    setTaskChecked,
    toggleTaskChecked,
    checkAll,
    uncheckAll,
    clear,
    reset,
    moveTask,
    getTaskIndex,
  }
})
