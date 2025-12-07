<script setup lang="ts">
  import { useRouter } from 'vue-router'
  import { Icon } from '@iconify/vue'
  import {
    ABOUT_ROUTE,
    HOME_ROUTE,
    LOGS_ROUTE,
    SETTINGS_ROUTE,
  } from '@/router/routes.ts'

  const router = useRouter()

  const navigateTo = (path: string) => {
    router.push(path)
  }

  const disableContextMenu = (event: MouseEvent) => {
    event.preventDefault()
    return false
  }

  const activeClass =
    'bg-blue-500 text-white dark:bg-blue-400 dark:text-gray-900'

  const defaultClass = 'text-gray-700 dark:text-gray-300'

  const hoverClass = 'hover:bg-gray-200 dark:hover:bg-gray-600'

  const isActiveRoute = (path: string) => {
    return router.currentRoute.value.path === path
  }

  const items = [
    {
      name: 'Home',
      icon: 'fluent:home-24-regular',
      route: HOME_ROUTE,
    },
    {
      name: 'Logs',
      icon: 'fluent:clock-24-regular',
      route: LOGS_ROUTE,
    },
    {
      name: 'About',  
      icon: 'fluent:info-24-regular',
      route: ABOUT_ROUTE,
    },
    {
      name: 'Settings',
      icon: 'fluent:settings-24-regular',
      route: SETTINGS_ROUTE,
    },
  ]
</script>

<template>
  <div
    :class="[
      'w-15 h-65 bg-white dark:bg-gray-800 rounded-2xl flex flex-col gap-y-5 px-2.5 py-5 shadow-2xl',
      'border border-gray-200 dark:border-gray-700',
      'transition-colors duration-250',
    ]"
    @contextmenu="disableContextMenu"
  >
    <div
      v-for="item in items"
      :key="item.route"
      :class="[
        'size-10 rounded-2xl transition-colors duration-250 ease-in-out',
        'flex items-center justify-center',
        isActiveRoute(item.route)
          ? activeClass
          : `${defaultClass} ${hoverClass}`,
      ]"
      @click="navigateTo(item.route)"
      :title="item.name"
    >
      <Icon :icon="item.icon" width="24" height="24" />
    </div>
  </div>
</template>

<style scoped></style>
