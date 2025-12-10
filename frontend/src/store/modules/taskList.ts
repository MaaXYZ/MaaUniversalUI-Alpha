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

/** Option value for a task (option name -> selected value) */
export type TaskOptionValues = Record<string, string>

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

  /** Currently selected task id for option configuration */
  const selectedTaskId = ref<string | null>(null)

  /** Option values for each task (task id -> option values) */
  const taskOptionValues = ref<Record<string, TaskOptionValues>>({})

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

  /** Currently selected task item */
  const selectedTask = computed(() => {
    if (!selectedTaskId.value) return null
    return taskList.value.find((item) => item.id === selectedTaskId.value) ?? null
  })

  /** Get option values for a specific task */
  const getTaskOptionValues = computed(() => {
    return (taskId: string): TaskOptionValues => {
      return taskOptionValues.value[taskId] ?? {}
    }
  })

  /** Get current selected task's option values */
  const selectedTaskOptionValues = computed(() => {
    if (!selectedTaskId.value) return {}
    return taskOptionValues.value[selectedTaskId.value] ?? {}
  })

  // ============ Config Sync ============

  /** Load task list from config (option sync is handled by backend) */
  function loadFromConfig() {
    const configTasks = configStore.tasks
    const piTasks = piStore.tasks

    // clear existing task list and option values
    taskList.value = []
    taskOptionValues.value = {}

    // restore task list from config
    for (const configTask of configTasks) {
      const piTask = piTasks.find((t: pi.V2Task) => t.name === configTask.name)
      if (piTask) {
        taskList.value.push({
          id: configTask.id,
          task: piTask,
          checked: configTask.checked,
        })

        // Restore option values from config (already synced by backend)
        if (configTask.option && configTask.option.length > 0) {
          const savedOptions: TaskOptionValues = {}
          for (const opt of configTask.option) {
            savedOptions[opt.name] = opt.value
          }
          taskOptionValues.value[configTask.id] = savedOptions
        }
      }
    }
  }

  /** Sync task list to config and save */
  async function syncToConfig() {
    // update task list in config
    configStore.setTasks(
      taskList.value.map((item) => {
        // Convert option values to ConfigTaskOption array
        const optionValues = taskOptionValues.value[item.id] ?? {}
        const options: pi.ConfigTaskOption[] = Object.entries(optionValues).map(
          ([name, value]) =>
            pi.ConfigTaskOption.createFrom({ name, value })
        )

        return pi.ConfigTask.createFrom({
          id: item.id,
          name: item.task.name,
          checked: item.checked,
          option: options,
        })
      })
    )

    // save to backend
    await configStore.save()
  }

  // ============ Actions ============

  /** Initialize task list (add default tasks according to default_check) */
  function initFromPi() {
    const defaultTasks = piStore.getDefaultCheckedTasks()
    taskList.value = defaultTasks.map((task) => {
      const id = generateUUID()
      return {
        id,
        task,
        checked: true,
      }
    })
  }

  /** Add task to list (same task can be added repeatedly) */
  function addTask(task: pi.V2Task, checked = false): string {
    const id = generateUUID()
    taskList.value.push({ id, task, checked })

    // Initialize default option values for the new task
    if (task.option && task.option.length > 0) {
      const defaults: TaskOptionValues = {}
      initOptionDefaults(defaults, task.option)
      taskOptionValues.value[id] = defaults
    }

    syncToConfig()
    return id
  }

  /** Add tasks in batch */
  function addTasks(tasks: pi.V2Task[], checked = false): string[] {
    const ids = tasks.map((task) => {
      const id = generateUUID()
      taskList.value.push({ id, task, checked })

      // Initialize default option values for each new task
      if (task.option && task.option.length > 0) {
        const defaults: TaskOptionValues = {}
        initOptionDefaults(defaults, task.option)
        taskOptionValues.value[id] = defaults
      }

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
      // Clear option values for removed task
      delete taskOptionValues.value[id]
      // Clear selection if removed task was selected
      if (selectedTaskId.value === id) {
        selectedTaskId.value = null
      }
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
    selectedTaskId.value = null
    taskOptionValues.value = {}
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

  /** Select a task for option configuration */
  function selectTask(id: string | null) {
    selectedTaskId.value = id
  }

  /** Set option value for a task */
  function setOptionValue(taskId: string, optionName: string, value: string, inputName?: string) {
    if (!taskOptionValues.value[taskId]) {
      taskOptionValues.value[taskId] = {}
    }
    const taskOpts = taskOptionValues.value[taskId]
    if (!taskOpts) return

    const key = inputName ? `${optionName}.${inputName}` : optionName
    const oldValue = taskOpts[key]
    taskOpts[key] = value

    // If this is a select/switch option, sync nested options for real-time UI update
    if (!inputName && oldValue !== value) {
      const option = piStore.getOptionByName(optionName)
      if (option && (option.type === 'select' || option.type === 'switch' || !option.type)) {
        // Remove old nested options
        const oldCase = option.cases?.find((c) => c.name === oldValue)
        if (oldCase?.option) {
          removeNestedOptions(taskOpts, oldCase.option)
        }

        // Add new nested options with defaults
        const newCase = option.cases?.find((c) => c.name === value)
        if (newCase?.option) {
          const defaults: TaskOptionValues = {}
          initOptionDefaults(defaults, newCase.option)
          for (const [k, v] of Object.entries(defaults)) {
            if (!(k in taskOpts)) {
              taskOpts[k] = v
            }
          }
        }
      }
    }

    syncToConfig()
  }

  /** Remove nested options recursively (for real-time UI update when switching options) */
  function removeNestedOptions(taskOpts: TaskOptionValues, optionNames: string[]) {
    for (const optionName of optionNames) {
      const option = piStore.getOptionByName(optionName)
      if (!option) continue

      if (option.type === 'input') {
        // Remove all input fields
        if (option.inputs) {
          for (const input of option.inputs) {
            delete taskOpts[`${optionName}.${input.name}`]
          }
        }
      } else {
        // Remove the option itself
        const currentValue = taskOpts[optionName]
        delete taskOpts[optionName]

        // Remove nested options of the currently selected case
        const currentCase = option.cases?.find((c) => c.name === currentValue)
        if (currentCase?.option) {
          removeNestedOptions(taskOpts, currentCase.option)
        }
      }
    }
  }

  /** Initialize option defaults (for real-time UI update when switching options) */
  function initOptionDefaults(values: TaskOptionValues, optionNames: string[]) {
    for (const optionName of optionNames) {
      const option = piStore.getOptionByName(optionName)
      if (!option) continue

      if (option.type === 'select' || !option.type) {
        let selectedCase: pi.V2OptionCase | undefined
        if (option.default_case) {
          values[optionName] = option.default_case
          selectedCase = option.cases?.find((c) => c.name === option.default_case)
        } else if (option.cases && option.cases.length > 0) {
          const firstCase = option.cases[0]
          if (firstCase) {
            values[optionName] = firstCase.name
            selectedCase = firstCase
          }
        } else {
          values[optionName] = ''
        }

        if (selectedCase?.option && selectedCase.option.length > 0) {
          initOptionDefaults(values, selectedCase.option)
        }
      } else if (option.type === 'switch') {
        const cases = option.cases ?? []
        const noCase = cases.find(
          (c) => !['Yes', 'yes', 'Y', 'y'].includes(c.name)
        )
        values[optionName] = noCase?.name ?? 'No'

        if (noCase?.option && noCase.option.length > 0) {
          initOptionDefaults(values, noCase.option)
        }
      } else if (option.type === 'input') {
        if (option.inputs) {
          for (const input of option.inputs) {
            let value = input.default ?? ''
            if (!value) {
              switch (input.pipeline_type) {
                case 'bool':
                  value = 'false'
                  break
                case 'int':
                  value = '0'
                  break
                default:
                  value = ''
              }
            }
            values[`${optionName}.${input.name}`] = value
          }
        }
      }
    }
  }

  /** Set multiple option values for a task */
  function setOptionValues(taskId: string, values: TaskOptionValues) {
    taskOptionValues.value[taskId] = { ...taskOptionValues.value[taskId], ...values }
    syncToConfig()
  }

  /** Get option value for a task */
  function getOptionValue(taskId: string, optionName: string): string | undefined {
    return taskOptionValues.value[taskId]?.[optionName]
  }

  return {
    // State
    taskList,
    selectedTaskId,
    taskOptionValues,

    // Getters
    availableTasks,
    checkedTasks,
    checkedItems,
    checkedCount,
    totalCount,
    hasAvailableTasks,
    selectedTask,
    getTaskOptionValues,
    selectedTaskOptionValues,

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
    selectTask,
    setOptionValue,
    setOptionValues,
    getOptionValue,
  }
})
