<script setup lang="ts">
import { XMarkdownContent } from '@novelia/web-kit';

import { getMarkdown, type MarkdownMode } from '@/markdown';

defineProps<{
  mode: MarkdownMode;
  source: string;
}>();
</script>

<template>
  <XMarkdownContent
    :source="source"
    :markdown="getMarkdown(mode)"
    :class="`forum-markdown forum-markdown--${mode}`"
  />
</template>

<style scoped src="./markdown.css"></style>
