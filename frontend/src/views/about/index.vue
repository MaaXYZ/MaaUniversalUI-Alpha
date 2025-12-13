<script setup lang="ts">
  import { Icon } from '@iconify/vue'
  import { usePiStore } from '@/store'
  import { GetMaaVersion } from '@wails/go/engine/service'
  import type { system } from '@wails/go/models'
  import { GetAppInfo } from '@wails/go/system/service'
  import { BrowserOpenURL } from '@wails/runtime/runtime'
  import { computed, onMounted, ref } from 'vue'

  const piStore = usePiStore()
  const maaVersion = ref('')
  const appInfo = ref<system.AppInfo>()
  const isLoading = ref(true)

  function formatISODateTimeToLocal(value?: string) {
    if (!value) return '未知'

    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value

    const pad2 = (n: number) => String(n).padStart(2, '0')
    return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(
      date.getDate()
    )} ${pad2(date.getHours())}:${pad2(date.getMinutes())}:${pad2(
      date.getSeconds()
    )}`
  }

  const projectVersion = computed(() => piStore.version || '未知')

  const buildInfoList = computed(() => [
    {
      label: '构建时间',
      value: formatISODateTimeToLocal(appInfo.value?.build_at),
      icon: 'fluent:calendar-20-regular',
    },
    {
      label: '构建系统',
      value: appInfo.value?.build_os || '未知',
      icon: 'fluent:desktop-20-regular',
    },
    {
      label: '构建架构',
      value: appInfo.value?.build_arch || '未知',
      icon: 'ri:cpu-line',
    },
    {
      label: 'Go 版本',
      value: appInfo.value?.go_version || '未知',
      icon: 'simple-icons:go',
    },
  ])

  const baseLinksList = [
    {
      name: 'MUU-Alpha',
      description: 'MAA统一界面',
      icon: 'mdi:github',
      url: 'https://github.com/MaaXYZ/MaaUniversalUI-Alpha',
    },
    {
      name: 'MaaFramework',
      description: '基于图像识别的自动化黑盒测试框架',
      icon: 'mdi:github',
      url: 'https://github.com/MaaXYZ/MaaFramework',
    },
  ]

  const linksList = computed(() => {
    const name = (piStore.name || '').trim()
    const github = (piStore.github || '').trim()

    if (!name || !github) return baseLinksList
    if (baseLinksList.some((l) => l.url === github)) return baseLinksList

    return [
      {
        name,
        description: '项目仓库',
        icon: 'mdi:github',
        url: github,
      },
      ...baseLinksList,
    ]
  })

  onMounted(async () => {
    try {
      const [maaVer, info] = await Promise.all([GetMaaVersion(), GetAppInfo()])
      maaVersion.value = maaVer
      appInfo.value = info
    } catch (error) {
      console.error('Failed to load app info:', error)
    } finally {
      isLoading.value = false
    }
  })

  function openLink(url: string) {
    BrowserOpenURL(url)
  }
</script>

<template>
  <div class="pl-24 space-y-6 max-w-4xl">
    <header class="animate-fade-in">
      <div class="flex items-center gap-4 mb-2">
        <div
          class="w-14 h-14 rounded-2xl bg-linear-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-lg shadow-emerald-500/25 transition-transform duration-300 hover:scale-105"
        >
          <Icon
            icon="fluent:info-20-regular"
            width="28"
            height="28"
            class="text-white"
          />
        </div>
        <div>
          <h1
            class="text-2xl font-bold text-gray-800 dark:text-gray-100 tracking-tight"
          >
            关于
          </h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            了解应用信息和相关资源
          </p>
        </div>
      </div>
    </header>

    <!-- 版本信息卡片 -->
    <section
      class="animate-fade-in-up"
      style="animation-delay: 0.1s"
    >
      <div
        class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-lg overflow-hidden transition-all duration-300 hover:shadow-xl"
      >
        <!-- 卡片头部 -->
        <div
          class="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80"
        >
          <div
            class="w-10 h-10 rounded-lg bg-linear-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-md"
          >
            <Icon
              icon="fluent:info-20-regular"
              width="22"
              height="22"
              class="text-white"
            />
          </div>
          <div class="flex-1">
            <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-100">
              版本信息
            </h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              当前应用和框架版本
            </p>
          </div>
        </div>

        <!-- 卡片内容 -->
        <div class="p-5">
          <!-- 加载状态 -->
          <div
            v-if="isLoading"
            class="flex items-center justify-center py-8"
          >
            <Icon
              icon="fluent:spinner-ios-20-regular"
              width="32"
              height="32"
              class="text-gray-400 animate-spin"
            />
          </div>

          <!-- 版本信息网格 -->
          <div
            v-else
            class="grid grid-cols-1 md:grid-cols-3 gap-4"
          >
            <!-- 应用版本 -->
            <div
              class="group p-4 rounded-lg bg-linear-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 border border-emerald-200 dark:border-emerald-800/50 transition-all duration-300 hover:scale-[1.02] hover:shadow-md"
            >
              <div class="flex items-center gap-2 mb-2">
                <Icon
                  icon="fluent:app-generic-20-regular"
                  width="18"
                  height="18"
                  class="text-emerald-600 dark:text-emerald-400"
                />
                <span
                  class="text-xs font-medium text-emerald-700 dark:text-emerald-300"
                >
                  应用版本
                </span>
              </div>
              <p
                class="text-lg font-bold text-emerald-800 dark:text-emerald-200 font-mono"
              >
                {{ appInfo?.version || '未知' }}
              </p>
            </div>

            <!-- 项目版本 -->
            <div
              class="group p-4 rounded-lg bg-linear-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 border border-blue-200 dark:border-blue-800/50 transition-all duration-300 hover:scale-[1.02] hover:shadow-md"
            >
              <div class="flex items-center gap-2 mb-2">
                <Icon
                  icon="fluent:folder-20-regular"
                  width="18"
                  height="18"
                  class="text-blue-600 dark:text-blue-400"
                />
                <span
                  class="text-xs font-medium text-blue-700 dark:text-blue-300"
                >
                  项目版本
                </span>
              </div>
              <p
                class="text-lg font-bold text-blue-800 dark:text-blue-200 font-mono"
              >
                {{ projectVersion }}
              </p>
            </div>

            <!-- MAA 版本 -->
            <div
              class="group p-4 rounded-lg bg-linear-to-br from-purple-50 to-violet-50 dark:from-purple-900/20 dark:to-violet-900/20 border border-purple-200 dark:border-purple-800/50 transition-all duration-300 hover:scale-[1.02] hover:shadow-md"
            >
              <div class="flex items-center gap-2 mb-2">
                <Icon
                  icon="fluent:bot-20-regular"
                  width="18"
                  height="18"
                  class="text-purple-600 dark:text-purple-400"
                />
                <span
                  class="text-xs font-medium text-purple-700 dark:text-purple-300"
                >
                  MAA 版本
                </span>
              </div>
              <p
                class="text-lg font-bold text-purple-800 dark:text-purple-200 font-mono"
              >
                {{ maaVersion || '未知' }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 构建信息卡片 -->
    <section
      class="animate-fade-in-up"
      style="animation-delay: 0.2s"
    >
      <div
        class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-lg overflow-hidden transition-all duration-300 hover:shadow-xl"
      >
        <!-- 卡片头部 -->
        <div
          class="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80"
        >
          <div
            class="w-10 h-10 rounded-lg bg-linear-to-br from-amber-500 to-orange-600 flex items-center justify-center shadow-md"
          >
            <Icon
              icon="fluent:wrench-20-regular"
              width="22"
              height="22"
              class="text-white"
            />
          </div>
          <div class="flex-1">
            <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-100">
              构建信息
            </h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              应用构建详细信息
            </p>
          </div>
        </div>

        <!-- 卡片内容 -->
        <div class="p-5">
          <div
            v-if="isLoading"
            class="flex items-center justify-center py-8"
          >
            <Icon
              icon="fluent:spinner-ios-20-regular"
              width="32"
              height="32"
              class="text-gray-400 animate-spin"
            />
          </div>

          <div
            v-else
            class="grid grid-cols-2 md:grid-cols-4 gap-4"
          >
            <div
              v-for="(item, index) in buildInfoList"
              :key="item.label"
              class="group p-4 rounded-lg bg-gray-50 dark:bg-gray-700/50 border border-gray-200 dark:border-gray-600 transition-all duration-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:scale-[1.02]"
              :style="{ animationDelay: `${0.3 + index * 0.05}s` }"
            >
              <div class="flex items-center gap-2 mb-2">
                <Icon
                  :icon="item.icon"
                  width="16"
                  height="16"
                  class="text-gray-500 dark:text-gray-400 group-hover:text-emerald-500 dark:group-hover:text-emerald-400 transition-colors"
                />
                <span class="text-xs text-gray-500 dark:text-gray-400">
                  {{ item.label }}
                </span>
              </div>
              <p
                class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate font-mono"
                :title="item.value"
              >
                {{ item.value }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 相关链接卡片 -->
    <section
      class="animate-fade-in-up"
      style="animation-delay: 0.4s"
    >
      <div
        class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-lg overflow-hidden transition-all duration-300 hover:shadow-xl"
      >
        <!-- 卡片头部 -->
        <div
          class="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/80"
        >
          <div
            class="w-10 h-10 rounded-lg bg-linear-to-br from-rose-500 to-pink-600 flex items-center justify-center shadow-md"
          >
            <Icon
              icon="fluent:link-20-regular"
              width="22"
              height="22"
              class="text-white"
            />
          </div>
          <div class="flex-1">
            <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-100">
              相关链接
            </h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              获取更多资源和帮助
            </p>
          </div>
        </div>

        <!-- 卡片内容 -->
        <div class="p-5">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <button
              v-for="(link, index) in linksList"
              :key="link.name"
              class="group p-4 rounded-lg bg-gray-50 dark:bg-gray-700/50 border border-gray-200 dark:border-gray-600 transition-all duration-300 hover:bg-gradient-to-br hover:from-emerald-50 hover:to-teal-50 dark:hover:from-emerald-900/20 dark:hover:to-teal-900/20 hover:border-emerald-300 dark:hover:border-emerald-700 hover:shadow-md hover:scale-[1.02] cursor-pointer text-left"
              :style="{ animationDelay: `${0.5 + index * 0.05}s` }"
              @click="openLink(link.url)"
            >
              <div class="flex items-start gap-3">
                <div
                  class="w-10 h-10 rounded-lg flex items-center justify-center bg-white dark:bg-gray-600 shadow-sm group-hover:shadow-md group-hover:bg-emerald-100 dark:group-hover:bg-emerald-800/50 transition-all"
                >
                  <Icon
                    :icon="link.icon"
                    width="22"
                    height="22"
                    class="text-gray-600 dark:text-gray-300 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors"
                  />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-1">
                    <h4
                      class="text-sm font-semibold text-gray-800 dark:text-gray-200 group-hover:text-emerald-700 dark:group-hover:text-emerald-300 transition-colors"
                    >
                      {{ link.name }}
                    </h4>
                    <Icon
                      icon="fluent:arrow-up-right-20-regular"
                      width="14"
                      height="14"
                      class="text-gray-400 group-hover:text-emerald-500 transition-all group-hover:translate-x-0.5 group-hover:-translate-y-0.5"
                    />
                  </div>
                  <p
                    class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-2"
                  >
                    {{ link.description }}
                  </p>
                </div>
              </div>
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
  /* 淡入动画 */
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  /* 淡入上移动画 */
  @keyframes fade-in-up {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .animate-fade-in {
    animation: fade-in 0.5s ease-out forwards;
  }

  .animate-fade-in-up {
    opacity: 0;
    animation: fade-in-up 0.5s ease-out forwards;
  }

  /* 行数限制 */
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
