<script setup lang="ts">
import { MoreHorizOutlined } from '@vicons/material';
import { XButton } from '@novelia/web-kit';
import { NDropdown, type DropdownOption } from 'naive-ui';
import { computed } from 'vue';

import type { Comment, CommentStatusName } from '@/api';

const props = defineProps<{ comment: Comment; actionsDisabled: boolean }>();
const emit = defineEmits<{
  filterPost: [postId: number];
  moderate: [comment: Comment, status: CommentStatusName];
}>();

const actionOptions = computed<DropdownOption[]>(() => {
  const options: DropdownOption[] = [{ label: '仅看此帖', key: 'filter-post' }];
  if (props.comment.status !== 0)
    options.push({
      label: '恢复',
      key: 'published',
      disabled: props.actionsDisabled,
    });
  if (props.comment.status !== 1)
    options.push({
      label: '隐藏',
      key: 'hidden',
      disabled: props.actionsDisabled,
    });
  if (props.comment.status !== 2)
    options.push({
      label: '删除',
      key: 'deleted',
      disabled: props.actionsDisabled,
    });
  return options;
});

function selectAction(key: string | number) {
  if (key === 'filter-post') {
    emit('filterPost', props.comment.postId);
    return;
  }
  if (props.actionsDisabled) return;
  emit('moderate', props.comment, key as CommentStatusName);
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}

function formatCommentId(comment: Comment) {
  return comment.rootId
    ? `#${comment.id} → #${comment.rootId}`
    : `#${comment.id}`;
}
</script>

<template>
  <div class="flex flex-col gap-3.5">
    <div class="flex items-start justify-between gap-3">
      <div class="flex min-w-0 flex-col gap-[3px]">
        <strong class="text-ink">{{ comment.authorUsername }}</strong>
        <div class="flex flex-wrap items-center gap-x-1 text-xs text-muted">
          <span>{{ formatCommentId(comment) }} ·</span>
          <a
            class="text-primary hover:underline"
            :href="`/p/${comment.postId}`"
            target="_blank"
            rel="noopener noreferrer"
          >
            帖子 #{{ comment.postId }}
          </a>
          <span>· {{ formatDate(comment.createdAt) }}</span>
        </div>
      </div>
      <div class="flex shrink-0 items-center gap-1.5">
        <span
          class="rounded px-2 py-0.5 text-xs"
          :class="
            comment.status === 0
              ? 'bg-success-soft text-success-text'
              : comment.status === 1
                ? 'bg-warning-soft text-warning-text'
                : 'bg-error-soft text-error-text'
          "
        >
          {{ ['正常发布', '隐藏', '删除'][comment.status] }}
        </span>
        <n-dropdown
          trigger="click"
          :options="actionOptions"
          @select="selectAction"
        >
          <XButton
            color="neutral"
            variant="ghost"
            size="icon"
            class="size-8"
            aria-label="评论操作"
          >
            <MoreHorizOutlined class="size-5" aria-hidden="true" />
          </XButton>
        </n-dropdown>
      </div>
    </div>
    <p
      class="whitespace-pre-wrap leading-[1.72] text-ink [overflow-wrap:anywhere]"
    >
      {{ comment.content }}
    </p>
  </div>
</template>
