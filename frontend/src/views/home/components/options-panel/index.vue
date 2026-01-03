<script setup lang="ts">
  import { computed } from 'vue'
  import { Icon } from '@iconify/vue'
  import { usePiStore, useTaskListStore } from '@/store/modules'
  import { pi } from '@wails/go/models'
  import { MarkdownContent } from '@/components'

  const piStore = usePiStore()
  const taskListStore = useTaskListStore()

  /** Currently selected task */
  const selectedTask = computed(() => taskListStore.selectedTask)

  /** Get task options configuration list */
  const taskOptions = computed(() => {
    if (!selectedTask.value?.task.option) return []
    return selectedTask.value.task.option
      .map((optName: string) => ({
        name: optName,
        option: piStore.getOptionByName(optName),
      }))
      .filter((item) => item.option !== null) as Array<{
      name: string
      option: pi.V2Option
    }>
  })

  /** Get display name for task */
  const taskDisplayName = computed(() => {
    if (!selectedTask.value) return ''
    const label = selectedTask.value.task.label
    if (label) {
      return piStore.resolveI18n(label)
    }
    return selectedTask.value.task.name
  })

  /** Check if task has options */
  const hasOptions = computed(() => taskOptions.value.length > 0)

  /** Get option value */
  function getOptionValue(optionName: string, inputName?: string): string {
    if (!selectedTask.value) return ''
    const key = inputName ? `${optionName}.${inputName}` : optionName
    return taskListStore.getOptionValue(selectedTask.value.id, key) ?? ''
  }

  /** Set option value */
  function setOptionValue(
    optionName: string,
    value: string,
    inputName?: string
  ) {
    if (!selectedTask.value) return
    taskListStore.setOptionValue(
      selectedTask.value.id,
      optionName,
      value,
      inputName
    )
  }

  /** Get option label */
  function getOptionLabel(option: pi.V2Option, name: string): string {
    if (option.label) {
      return piStore.resolveI18n(option.label)
    }
    return name
  }

  /** Get case label */
  function getCaseLabel(optCase: pi.V2OptionCase): string {
    if (optCase.label) {
      return piStore.resolveI18n(optCase.label)
    }
    return optCase.name
  }

  /** Get input label */
  function getInputLabel(input: pi.V2OptionInput): string {
    if (input.label) {
      return piStore.resolveI18n(input.label)
    }
    return input.name
  }

  /** Check if switch value is Yes */
  function isSwitchYes(optionName: string): boolean {
    const value = getOptionValue(optionName)
    return ['Yes', 'yes', 'Y', 'y'].includes(value)
  }

  /** Toggle switch value */
  function toggleSwitch(optionName: string, option: pi.V2Option) {
    const cases = option.cases ?? []
    const yesCase = cases.find((c) => ['Yes', 'yes', 'Y', 'y'].includes(c.name))
    const noCase = cases.find((c) => !['Yes', 'yes', 'Y', 'y'].includes(c.name))

    const currentValue = getOptionValue(optionName)
    const isCurrentYes = ['Yes', 'yes', 'Y', 'y'].includes(currentValue)

    if (isCurrentYes && noCase) {
      setOptionValue(optionName, noCase.name)
    } else if (!isCurrentYes && yesCase) {
      setOptionValue(optionName, yesCase.name)
    }
  }

  /** Validate input value */
  function validateInput(input: pi.V2OptionInput, value: string): boolean {
    if (!input.verify) return true
    try {
      const regex = new RegExp(input.verify)
      return regex.test(value)
    } catch {
      return true
    }
  }

  /** Close panel */
  function closePanel() {
    taskListStore.selectTask(null)
  }
</script>

<template>
  <div
    class="w-87.5 h-full rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col shadow-lg"
  >
    <!-- Header -->
    <div
      class="h-13 flex justify-between items-center rounded-t-xl border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80 px-5 py-4"
    >
      <div class="flex items-center gap-2">
        <Icon
          icon="fluent:settings-20-regular"
          width="20"
          height="20"
          class="text-indigo-500"
        />
        <span class="text-sm font-semibold text-gray-700 dark:text-gray-200">
          配置选项
        </span>
      </div>

      <!-- Close button when task is selected -->
      <button
        v-if="selectedTask"
        class="p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors duration-200 cursor-pointer"
        @click="closePanel"
      >
        <Icon
          icon="fluent:dismiss-20-regular"
          width="18"
          height="18"
        />
      </button>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-y-auto p-4 space-y-4 scrollbar-none">
      <!-- No task selected state -->
      <div
        v-if="!selectedTask"
        class="flex flex-col items-center justify-center h-full text-gray-400 dark:text-gray-500"
      >
        <Icon
          icon="fluent:cursor-click-20-regular"
          width="48"
          height="48"
          class="mb-3 opacity-50"
        />
        <p class="text-sm">点击左侧任务进行配置</p>
      </div>

      <!-- Task selected but no options -->
      <div
        v-else-if="!hasOptions"
        class="flex flex-col items-center justify-center h-full text-gray-400 dark:text-gray-500"
      >
        <Icon
          icon="fluent:checkbox-checked-20-regular"
          width="48"
          height="48"
          class="mb-3 opacity-50"
        />
        <p class="text-sm font-medium text-gray-600 dark:text-gray-300 mb-1">
          {{ taskDisplayName }}
        </p>
        <p class="text-sm">该任务无需配置</p>
      </div>

      <!-- Options form -->
      <template v-else>
        <!-- Task name header -->
        <div class="pb-3 border-b border-gray-100 dark:border-gray-700">
          <h3 class="text-base font-semibold text-gray-800 dark:text-gray-100">
            {{ taskDisplayName }}
          </h3>
          <MarkdownContent
            v-if="selectedTask.task.description"
            :content="selectedTask.task.description"
            class="mt-1 text-xs"
          />
        </div>

        <!-- Option items -->
        <div
          v-for="{ name, option } in taskOptions"
          :key="name"
          class="space-y-2"
        >
          <!-- Option label -->
          <label
            class="block text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            {{ getOptionLabel(option, name) }}
          </label>

          <!-- Option description -->
          <MarkdownContent
            v-if="option.description"
            :content="option.description"
            class="text-xs"
          />

          <!-- Select type -->
          <div
            v-if="option.type === 'select' || !option.type"
            class="relative"
          >
            <select
              :value="getOptionValue(name)"
              class="w-full px-3 py-2 text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent appearance-none cursor-pointer pr-10"
              @change="
                setOptionValue(name, ($event.target as HTMLSelectElement).value)
              "
            >
              <option
                v-for="optCase in option.cases"
                :key="optCase.name"
                :value="optCase.name"
              >
                {{ getCaseLabel(optCase) }}
              </option>
            </select>
            <Icon
              icon="fluent:chevron-down-20-regular"
              width="18"
              height="18"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"
            />
          </div>

          <!-- Switch type -->
          <div
            v-else-if="option.type === 'switch'"
            class="flex items-center gap-3"
          >
            <button
              class="relative w-11 h-6 rounded-full transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
              :class="
                isSwitchYes(name)
                  ? 'bg-indigo-500'
                  : 'bg-gray-300 dark:bg-gray-600'
              "
              @click="toggleSwitch(name, option)"
            >
              <span
                class="absolute top-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200"
                :class="isSwitchYes(name) ? 'left-5.5' : 'left-0.5'"
              />
            </button>
            <span class="text-sm text-gray-600 dark:text-gray-300">
              {{ isSwitchYes(name) ? '开启' : '关闭' }}
            </span>
          </div>

          <!-- Input type -->
          <div
            v-else-if="option.type === 'input'"
            class="space-y-3"
          >
            <div
              v-for="input in option.inputs"
              :key="input.name"
              class="space-y-1"
            >
              <label
                class="block text-xs font-medium text-gray-600 dark:text-gray-400"
              >
                {{ getInputLabel(input) }}
              </label>
              <MarkdownContent
                v-if="input.description"
                :content="input.description"
                class="text-xs"
              />
              <input
                type="text"
                :value="getOptionValue(name, input.name)"
                :placeholder="input.default"
                class="w-full px-3 py-2 text-sm bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                :class="{
                  'border-red-500! focus:ring-red-500!':
                    getOptionValue(name, input.name) &&
                    !validateInput(input, getOptionValue(name, input.name)),
                }"
                @input="
                  setOptionValue(
                    name,
                    ($event.target as HTMLInputElement).value,
                    input.name
                  )
                "
              />
              <p
                v-if="
                  getOptionValue(name, input.name) &&
                  !validateInput(input, getOptionValue(name, input.name)) &&
                  input.pattern_msg
                "
                class="text-xs text-red-500"
              >
                {{ piStore.resolveI18n(input.pattern_msg) }}
              </p>
            </div>
          </div>

          <!-- Divider between options -->
          <div class="pt-2" />
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped></style>
