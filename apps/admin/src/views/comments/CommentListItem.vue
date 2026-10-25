<script setup lang="ts">
import { MoreHorizOutlined } from '@vicons/material';
import {
  XButton,
  XDropdownMenu,
  XDropdownMenuItem,
  XDropdownMenuSeparator,
} from '@novelia/web-kit';

import type { Comment, CommentStatusName } from '@/api';

const props = defineProps<{ comment: Comment; actionsDisabled: boolean }>();
const emit = defineEmits<{
  filterSubject: [comment: Comment];
  moderate: [comment: Comment, status: CommentStatusName];
  deleteAuthor: [comment: Comment];
}>();

function selectAction(
  key: 'filter-subject' | 'delete-author' | CommentStatusName,
) {
  if (key === 'filter-subject') {
    emit('filterSubject', props.comment);
    return;
  }
  if (props.actionsDisabled) return;
  if (key === 'delete-author') {
    emit('deleteAuthor', props.comment);
    return;
  }
  emit('moderate', props.comment, key);
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
            v-if="comment.subjectType === 0 && comment.postId"
            class="text-primary hover:underline"
            :href="`/p/${comment.postId}`"
            target="_blank"
            rel="noopener noreferrer"
          >
            帖子 #{{ comment.postId }}
          </a>
          <span v-else class="break-all">小说 · {{ comment.subjectKey }}</span>
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
        <XDropdownMenu>
          <template #trigger>
            <XButton
              color="neutral"
              variant="ghost"
              size="icon-sm"
              aria-label="评论操作"
            >
              <MoreHorizOutlined class="size-5" aria-hidden="true" />
            </XButton>
          </template>
          <XDropdownMenuItem @select="selectAction('filter-subject')">
            仅看此资源
          </XDropdownMenuItem>
          <XDropdownMenuItem
            v-if="comment.status !== 0"
            :disabled="actionsDisabled"
            @select="selectAction('published')"
          >
            恢复
          </XDropdownMenuItem>
          <XDropdownMenuItem
            v-if="comment.status !== 1"
            :disabled="actionsDisabled"
            @select="selectAction('hidden')"
          >
            隐藏
          </XDropdownMenuItem>
          <XDropdownMenuItem
            v-if="comment.status !== 2"
            :disabled="actionsDisabled"
            destructive
            @select="selectAction('deleted')"
          >
            删除
          </XDropdownMenuItem>
          <XDropdownMenuSeparator />
          <XDropdownMenuItem
            :disabled="actionsDisabled"
            destructive
            @select="selectAction('delete-author')"
          >
            删除作者全部评论
          </XDropdownMenuItem>
        </XDropdownMenu>
      </div>
    </div>
    <p
      class="whitespace-pre-wrap leading-[1.72] text-ink [overflow-wrap:anywhere]"
    >
      {{ comment.content }}
    </p>
  </div>
</template>
