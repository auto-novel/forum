<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue';

import type { PostTag } from '@/api';
import AppButton from '@/components/AppButton.vue';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';

const MarkdownHelpDialog = defineAsyncComponent(
  () => import('@/components/markdown/MarkdownHelpDialog.vue'),
);

const props = withDefaults(
  defineProps<{
    categories: { id: number; slug: string; title: string }[];
    tags: PostTag[];
    submitting: boolean;
    submitLabel: string;
    submittingLabel: string;
    titlePlaceholder?: string;
    contentPlaceholder?: string;
    showTitleCount?: boolean;
    showCancel?: boolean;
  }>(),
  {
    titlePlaceholder: '',
    contentPlaceholder: '使用 Markdown 输入内容…',
    showTitleCount: false,
    showCancel: false,
  },
);

const emit = defineEmits<{
  submit: [];
  cancel: [];
  categoryChange: [];
}>();
const title = defineModel<string>('title', { required: true });
const category = defineModel<string>('category', { required: true });
const content = defineModel<string>('content', { required: true });
const tagIds = defineModel<number[]>('tagIds', { required: true });
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

    <div>
      <label
        for="post-category"
        class="mb-2 block text-sm font-semibold text-ink"
      >
        分类
      </label>
      <select
        id="post-category"
        v-model="category"
        class="block min-h-10 w-full rounded-md border border-border bg-transparent px-3 text-sm text-ink outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15 disabled:cursor-not-allowed disabled:opacity-50"
        :disabled="submitting"
        required
        @change="emit('categoryChange')"
      >
        <option
          v-for="categoryOption in categories"
          :key="categoryOption.id"
          :value="categoryOption.slug"
        >
          {{ categoryOption.title }}
        </option>
      </select>
    </div>

    <fieldset :disabled="submitting">
      <legend class="mb-2 text-sm font-semibold text-ink">标签</legend>
      <div v-if="tags.length" class="flex flex-wrap gap-2">
        <label
          v-for="tag in tags"
          :key="tag.id"
          class="cursor-pointer rounded-sm border px-3 py-1.5 text-xs font-medium transition-colors"
          :class="
            tagIds.includes(tag.id)
              ? 'border-primary bg-primary-soft text-primary'
              : 'border-border text-muted hover:border-primary hover:text-primary'
          "
        >
          <input
            v-model="tagIds"
            type="checkbox"
            class="sr-only"
            :value="tag.id"
          />
          {{ tag.name }}
        </label>
      </div>
      <p v-else class="text-sm text-muted">这个分类暂时没有可用标签。</p>
    </fieldset>

    <div>
      <div class="mb-2 flex items-center justify-between gap-3">
        <label class="text-sm font-semibold text-ink">正文</label>
        <MarkdownHelpDialog mode="article" />
      </div>
      <MarkdownEditor
        v-model="content"
        mode="article"
        :placeholder="contentPlaceholder"
        :rows="14"
        :disabled="submitting"
      />
    </div>

    <div
      class="flex flex-wrap items-center justify-end gap-3 border-t border-divider pt-5"
    >
      <div class="flex items-center gap-3">
        <AppButton
          v-if="showCancel"
          variant="outline"
          :disabled="submitting"
          @click="emit('cancel')"
        >
          取消
        </AppButton>
        <AppButton type="submit" :disabled="!canSubmit">
          {{ submitting ? submittingLabel : submitLabel }}
        </AppButton>
      </div>
    </div>
  </form>
</template>
