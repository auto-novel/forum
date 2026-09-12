<script setup lang="ts">
import { computed, useSlots } from 'vue';

import type { PostTag } from '@/api';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';
import TagSelector from '@/components/TagSelector.vue';

const props = withDefaults(
  defineProps<{
    tags: PostTag[];
    submitting: boolean;
    submitLabel: string;
    submittingLabel: string;
    tagLabel?: string;
    titlePlaceholder?: string;
    contentPlaceholder?: string;
    showTitleCount?: boolean;
    showCancel?: boolean;
  }>(),
  {
    titlePlaceholder: '',
    contentPlaceholder: '使用 Markdown 输入内容…',
    tagLabel: '标签（可选）',
    showTitleCount: false,
    showCancel: false,
  },
);

const emit = defineEmits<{
  submit: [];
  cancel: [];
}>();
const title = defineModel<string>('title', { required: true });
const content = defineModel<string>('content', { required: true });
const tagIds = defineModel<number[]>('tagIds', { required: true });
const slots = useSlots();

const canSubmit = computed(
  () =>
    Boolean(title.value.trim() && content.value.trim()) && !props.submitting,
);

function submit() {
  if (canSubmit.value) emit('submit');
}
</script>

<template>
  <form class="space-y-5" @submit.prevent="submit">
    <div>
      <label for="post-title" class="mb-2 block text-sm font-semibold text-ink">
        标题
      </label>
      <input
        id="post-title"
        v-model="title"
        type="text"
        maxlength="500"
        class="block min-h-10 w-full rounded-md border border-border bg-transparent px-3 text-sm text-ink outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15 disabled:cursor-not-allowed disabled:opacity-50"
        :placeholder="titlePlaceholder"
        :disabled="submitting"
        required
      />
      <p v-if="showTitleCount" class="mt-1 text-right text-xs text-muted">
        {{ title.length }} / 500
      </p>
    </div>

    <slot name="category" />

    <TagSelector
      v-model="tagIds"
      :tags="tags"
      :label="tagLabel"
      :disabled="submitting"
    />

    <div>
      <label class="mb-2 block text-sm font-semibold text-ink">正文</label>
      <MarkdownEditor
        v-model="content"
        mode="article"
        :placeholder="contentPlaceholder"
        :rows="14"
        :disabled="submitting"
      />
    </div>

    <div
      class="flex flex-wrap items-center gap-3 border-t border-divider pt-5"
      :class="slots.hint ? 'justify-between' : 'justify-end'"
    >
      <slot name="hint" />
      <div class="flex items-center gap-3">
        <button
          v-if="showCancel"
          type="button"
          class="rounded-sm border border-border px-4 py-2 text-sm font-medium hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="submitting"
          @click="emit('cancel')"
        >
          取消
        </button>
        <button
          type="submit"
          class="rounded-sm bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="!canSubmit"
        >
          {{ submitting ? submittingLabel : submitLabel }}
        </button>
      </div>
    </div>
  </form>
</template>
