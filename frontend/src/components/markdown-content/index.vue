<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { marked } from 'marked'
  import { ReadContent } from '@wails/go/pi/service'
  import { usePiStore } from '@/store/modules'

  const props = defineProps<{
    /** Content: can be file path, URL, or direct text */
    content?: string
    /** CSS class for the container */
    class?: string
  }>()

  const piStore = usePiStore()
  const renderedContent = ref('')
  const loading = ref(false)

  // Configure marked options
  marked.setOptions({
    breaks: true,
    gfm: true,
  })

  /** Load and render content */
  async function loadContent() {
    if (!props.content) {
      renderedContent.value = ''
      return
    }

    // First resolve i18n
    const resolved = piStore.resolveI18n(props.content)

    loading.value = true
    try {
      // Use backend to read content (handles file paths, URLs, and direct text)
      const content = await ReadContent(resolved)
      // Parse markdown to HTML
      renderedContent.value = await marked.parse(content)
    } catch (e) {
      console.error('Failed to load content:', e)
      // Fallback to direct text
      renderedContent.value = await marked.parse(resolved)
    } finally {
      loading.value = false
    }
  }

  // Watch for content changes
  watch(() => props.content, loadContent, { immediate: true })

  onMounted(loadContent)
</script>

<template>
  <div
    v-if="renderedContent"
    class="markdown-content"
    :class="props.class"
    v-html="renderedContent"
  />
  <div
    v-else-if="loading"
    class="text-gray-400 dark:text-gray-500 text-sm"
  >
    加载中...
  </div>
</template>

<style scoped>
  @reference "tailwindcss";

  .markdown-content {
    @apply text-sm text-gray-600 dark:text-gray-400 leading-relaxed;
  }

  .markdown-content :deep(p) {
    @apply mb-2 last:mb-0;
  }

  .markdown-content :deep(a) {
    @apply text-indigo-500 hover:text-indigo-600 dark:text-indigo-400 dark:hover:text-indigo-300 underline;
  }

  .markdown-content :deep(strong) {
    @apply font-semibold text-gray-700 dark:text-gray-300;
  }

  .markdown-content :deep(em) {
    @apply italic;
  }

  .markdown-content :deep(code) {
    @apply px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-xs font-mono;
  }

  .markdown-content :deep(pre) {
    @apply p-3 bg-gray-100 dark:bg-gray-700 rounded-lg overflow-x-auto mb-2;
  }

  .markdown-content :deep(pre code) {
    @apply p-0 bg-transparent;
  }

  .markdown-content :deep(ul),
  .markdown-content :deep(ol) {
    @apply pl-5 mb-2;
  }

  .markdown-content :deep(ul) {
    @apply list-disc;
  }

  .markdown-content :deep(ol) {
    @apply list-decimal;
  }

  .markdown-content :deep(li) {
    @apply mb-1;
  }

  .markdown-content :deep(blockquote) {
    @apply pl-4 border-l-4 border-gray-300 dark:border-gray-600 italic text-gray-500 dark:text-gray-400 mb-2;
  }

  .markdown-content :deep(h1),
  .markdown-content :deep(h2),
  .markdown-content :deep(h3),
  .markdown-content :deep(h4),
  .markdown-content :deep(h5),
  .markdown-content :deep(h6) {
    @apply font-semibold text-gray-800 dark:text-gray-200 mb-2;
  }

  .markdown-content :deep(h1) {
    @apply text-lg;
  }

  .markdown-content :deep(h2) {
    @apply text-base;
  }

  .markdown-content :deep(h3),
  .markdown-content :deep(h4),
  .markdown-content :deep(h5),
  .markdown-content :deep(h6) {
    @apply text-sm;
  }

  .markdown-content :deep(hr) {
    @apply border-gray-200 dark:border-gray-700 my-3;
  }

  .markdown-content :deep(table) {
    @apply w-full border-collapse mb-2;
  }

  .markdown-content :deep(th),
  .markdown-content :deep(td) {
    @apply border border-gray-200 dark:border-gray-700 px-3 py-1.5 text-left;
  }

  .markdown-content :deep(th) {
    @apply bg-gray-50 dark:bg-gray-800 font-semibold;
  }
</style>
