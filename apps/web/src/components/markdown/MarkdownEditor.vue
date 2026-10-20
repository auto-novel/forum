<script setup lang="ts">
import {
  StarBorderOutlined,
  UnfoldMoreOutlined,
  VisibilityOffOutlined,
} from '@vicons/material';
import { useTemplateRef } from 'vue';

import { XMarkdownEditor, XMarkdownEditorToolButton } from '@novelia/web-kit';

import type { MarkdownMode } from '@novelia/forum-api';
import MarkdownContent from './MarkdownContent.vue';

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
    class="rounded-md"
    :tools="['bold', 'italic', 'strike', 'link']"
    :sticky="mode === 'article'"
    :sticky-offset="50"
    :rows="rows"
    :maxlength="maxlength"
    :placeholder="placeholder"
    :disabled="disabled"
    :invalid="invalid"
    :aria-describedby="describedBy"
    spellcheck="false"
    :ui="{
      toolbar: 'rounded-t-md',
      textarea: 'min-h-24 placeholder:text-muted/70',
      preview: 'min-h-48',
    }"
  >
    <template #toolbar="{ wrapSelection, insertBlock, disabled: blocked }">
      <XMarkdownEditorToolButton
        label="剧透"
        :disabled="blocked"
        @click="wrapSelection('!!', '!!', '剧透内容')"
      >
        <VisibilityOffOutlined class="size-4" aria-hidden="true" />
      </XMarkdownEditorToolButton>
      <XMarkdownEditorToolButton
        label="评分"
        :disabled="blocked"
        @click="insertBlock('::: star 5', '', '')"
      >
        <StarBorderOutlined class="size-4" aria-hidden="true" />
      </XMarkdownEditorToolButton>
      <XMarkdownEditorToolButton
        label="折叠内容"
        :disabled="blocked"
        @click="insertBlock('::: details 点击展开\n', '\n:::', '折叠内容')"
      >
        <UnfoldMoreOutlined class="size-4" aria-hidden="true" />
      </XMarkdownEditorToolButton>
    </template>
    <template #preview="{ source }">
      <MarkdownContent v-if="source.trim()" :mode="mode" :source="source" />
      <p v-else class="text-sm text-muted">没有可预览的内容</p>
    </template>
  </XMarkdownEditor>
</template>
