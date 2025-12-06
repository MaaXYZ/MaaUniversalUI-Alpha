import type { RouteRecordRaw } from 'vue-router'

export const HOME_NAME = 'home'
export const LOGS_NAME = 'logs'
export const ABOUT_NAME = 'about'
export const SETTINGS_NAME = 'settings'

export const HOME_ROUTE = '/home'
export const LOGS_ROUTE = '/logs'
export const ABOUT_ROUTE = '/about'
export const SETTINGS_ROUTE = '/settings'

const routers: RouteRecordRaw[] = [
  {
    path: "/",
    component: () => import("@/layout/main/index.vue"),
    redirect: HOME_ROUTE,
    children: [
      {
        name: HOME_NAME,
        path: HOME_ROUTE,
        component: () => import("@/views/home/index.vue"),
      },
      {
        name: LOGS_NAME,
        path: LOGS_ROUTE, 
        component: () => import("@/views/logs/index.vue"),
      },
      {
        name: ABOUT_NAME,
        path: ABOUT_ROUTE,
        component: () => import("@/views/about/index.vue"),
      },
      {
        name: SETTINGS_NAME,
        path: SETTINGS_ROUTE,
        component: () => import("@/views/settings/index.vue"),
      },
    ]
  }
]

export default routers