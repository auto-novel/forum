<script setup lang="ts">
import { useTemplateRef } from 'vue';

import { XMarkdownEditor } from '@novelia/web-kit';

import { getMarkdown, type MarkdownMode } from '@/markdown';

withDefaults(
  defineProps<{
    mode: MarkdownMode;
    placeholder?: string;
    disabled?: boolean;
    describedBy?: string;
    invalid?: boolean;
    maxlength?: number;
    rows?: number;
  }>(),
  {
    placeholder: '使用 Markdown 输入内容…',
    disabled: false,
    maxlength: 100000,
    rows: 7,
  },
);

const value = defineModel<string>({ required: true });
const editor = useTemplateRef('editor');

defineExpose({ focus: () => editor.value?.focus() });
</script>

<template>
  <XMarkdownEditor
    ref="editor"
    v-model="value"
    :class="`forum-markdown forum-markdown--${mode}`"
    :markdown="getMarkdown(mode)"
    :tools="['bold', 'italic', 'strike', 'link', 'spoiler', 'star', 'details']"
    :sticky="mode === 'article'"
    :sticky-offset="50"
    :rows="rows"
    :maxlength="maxlength"
    :placeholder="placeholder"
    :disabled="disabled"
    :invalid="invalid"
    :aria-describedby="describedBy"
    spellcheck="false"
  />
</template>

<style scoped src="./markdown.css"></style>
