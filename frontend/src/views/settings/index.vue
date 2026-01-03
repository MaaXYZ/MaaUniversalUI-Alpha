<script setup lang="ts">
  import { ref, onMounted, onUnmounted, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import {
    ResSettingsCard,
    CtrlSettingsCard,
    ThemeSettingsCard,
    GeneralSettingsCard,
    SettingsSidebar,
  } from './components'

  const { t } = useI18n()

  const navItems = computed(() => [
    {
      id: 'general',
      label: t('settings.general.title'),
      icon: 'fluent:settings-20-regular',
    },
    {
      id: 'theme',
      label: t('settings.theme.title'),
      icon: 'fluent:color-20-regular',
    },
    {
      id: 'resource',
      label: t('settings.resource.title'),
      icon: 'fluent:folder-open-20-regular',
    },
    {
      id: 'controller',
      label: t('settings.controller.title'),
      icon: 'fluent:phone-link-setup-20-regular',
    },
  ])

  const activeId = ref('general')
  let isClickScrolling = false
  let observer: IntersectionObserver | null = null
  const ratios = new Map<string, number>()

  function handleSelect(id: string) {
    activeId.value = id
    isClickScrolling = true

    const el = document.getElementById(id)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' })
      // Debounce to prevent observer from overriding activeId during scroll
      setTimeout(() => {
        isClickScrolling = false
      }, 1000)
    }
  }

  onMounted(() => {
    observer = new IntersectionObserver(
      (entries) => {
        if (isClickScrolling) return

        entries.forEach((entry) => {
          ratios.set(
            entry.target.id,
            entry.isIntersecting ? entry.intersectionRatio : 0
          )
        })

        let maxRatio = 0
        let maxId = ''

        for (const item of navItems.value) {
          const ratio = ratios.get(item.id) || 0
          if (ratio > maxRatio) {
            maxRatio = ratio
            maxId = item.id
          }
        }

        if (maxId && maxRatio > 0) {
          activeId.value = maxId
        }
      },
      {
        root: null,
        rootMargin: '-10% 0px -80% 0px', // Narrow detection area near top
        threshold: [0, 0.2, 0.4, 0.6, 0.8, 1],
      }
    )

    navItems.value.forEach((item) => {
      const el = document.getElementById(item.id)
      if (el) observer?.observe(el)
    })
  })

  onUnmounted(() => {
    observer?.disconnect()
  })
</script>

<template>
  <div class="flex gap-8 pl-24 pt-2 relative h-full">
    <!-- Sidebar -->
    <aside class="w-56 shrink-0 hidden md:block">
      <div class="sticky top-6">
        <h2 class="text-lg font-bold text-gray-900 dark:text-white px-4 mb-4">
          {{ t('settings.title') }}
        </h2>
        <SettingsSidebar
          :items="navItems"
          :active-id="activeId"
          @select="handleSelect"
        />  
      </div>
    </aside>

    <!-- Content -->
    <main class="flex-1 min-w-0 max-w-3xl space-y-10 h-full overflow-y-auto scrollbar-none">
      <section
        id="general"
        class="scroll-mt-6"
      >
        <GeneralSettingsCard />
      </section>

      <section
        id="theme"
        class="scroll-mt-6"
      >
        <ThemeSettingsCard />
      </section>

      <section
        id="resource"
        class="scroll-mt-6"
      >
        <ResSettingsCard />
      </section>

      <section
        id="controller"
        class="scroll-mt-6"
      >
        <CtrlSettingsCard />
      </section>
    </main>
  </div>
</template>
