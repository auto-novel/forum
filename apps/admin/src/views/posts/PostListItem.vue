<script setup lang="ts">
import {
  CommentOutlined,
  LockOutlined,
  PushPinOutlined,
  VisibilityOutlined,
} from '@vicons/material';
import { XButton } from '@novelia/web-kit';
import { NInputNumber, NPopconfirm, NPopover, NSpin } from 'naive-ui';
import { ref, watch } from 'vue';

import type { PostStatus, PostSummary } from '@/api';

const props = defineProps<{
  post: PostSummary;
  categoryName: string;
  actionsDisabled: boolean;
  saving: boolean;
}>();

const emit = defineEmits<{
  setStatus: [post: PostSummary, status: PostStatus];
  setCommentsLocked: [post: PostSummary, locked: boolean];
  setPinOrder: [post: PostSummary, pinOrder: number | null];
  reviewComments: [post: PostSummary];
}>();

const pinOrderInput = ref<number | null>(props.post.pinOrder ?? null);
watch(
  () => props.post.pinOrder,
  (value) => (pinOrderInput.value = value ?? null),
);

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
</script>

<template>
  <div class="post-row">
    <div class="post-content">
      <div class="post-badges text-xs">
        <span class="rounded-sm bg-paper px-2 py-0.5 text-muted">
          {{ categoryName }}
        </span>
        <span
          v-if="post.status === 1"
          class="rounded-sm bg-warning-soft px-2 py-0.5 text-warning-text"
        >
          隐藏
        </span>
        <span
          v-if="post.status === 2"
          class="rounded-sm bg-error-soft px-2 py-0.5 text-error-text"
        >
          已删除
        </span>
        <span
          v-if="post.pinOrder != null"
          class="inline-flex items-center gap-1 rounded-sm bg-warning-soft px-2 py-0.5 text-warning-text"
        >
          <PushPinOutlined class="size-3.5" aria-hidden="true" />
          置顶 {{ post.pinOrder }}
        </span>
        <span
          v-if="post.commentsLocked"
          class="inline-flex items-center gap-1 rounded-sm bg-error-soft px-2 py-0.5 text-error-text"
        >
          <LockOutlined class="size-3.5" aria-hidden="true" />
          已锁评
        </span>
      </div>
      <p class="post-title font-semibold text-ink">{{ post.title }}</p>
      <div class="post-footer">
        <span class="text-muted">
          {{ post.authorUsername }} (ID {{ post.authorId }}) ·
          {{ formatDate(post.activeAt) }}
        </span>
        <div class="post-metrics">
          <span>
            <VisibilityOutlined class="size-4" aria-hidden="true" />
            {{ post.viewsCount }}
          </span>
          <span>
            <CommentOutlined class="size-4" aria-hidden="true" />
            {{ post.commentsCount }}
          </span>
          <XButton
            variant="ghost"
            color="primary"
            size="sm"
            @click="emit('reviewComments', post)"
          >
            审核评论
          </XButton>
        </div>
      </div>
      <div class="post-actions">
        <div class="action-group">
          <XButton
            color="neutral"
            variant="ghost"
            size="sm"
            :disabled="actionsDisabled"
            @click="emit('setCommentsLocked', post, !post.commentsLocked)"
          >
            {{ post.commentsLocked ? '解锁' : '锁定' }}
          </XButton>
          <n-popover
            trigger="click"
            placement="bottom-start"
            :disabled="actionsDisabled"
          >
            <template #trigger>
              <XButton
                color="neutral"
                variant="ghost"
                size="sm"
                :disabled="actionsDisabled"
              >
                {{ post.pinOrder == null ? '置顶' : '置顶设置' }}
              </XButton>
            </template>
            <div class="pin-settings">
              <p class="font-semibold text-ink">置顶设置</p>
              <label :for="`pin-order-${post.id}`" class="action-label">
                置顶顺序
              </label>
              <n-input-number
                v-model:value="pinOrderInput"
                :input-props="{ id: `pin-order-${post.id}` }"
                size="small"
                :min="-2147483648"
                :max="2147483647"
                :precision="0"
                :show-button="false"
                :disabled="actionsDisabled"
                placeholder="输入顺序"
              />
              <div class="pin-actions">
                <XButton
                  v-if="post.pinOrder != null"
                  color="neutral"
                  variant="ghost"
                  size="sm"
                  :disabled="actionsDisabled"
                  @click="emit('setPinOrder', post, null)"
                >
                  取消置顶
                </XButton>
                <XButton
                  variant="soft"
                  color="primary"
                  size="sm"
                  :disabled="
                    actionsDisabled ||
                    pinOrderInput == null ||
                    pinOrderInput === (post.pinOrder ?? null)
                  "
                  @click="emit('setPinOrder', post, pinOrderInput)"
                >
                  保存
                </XButton>
              </div>
            </div>
          </n-popover>
        </div>
        <div class="action-group status-actions">
          <n-spin v-if="saving" size="small" />
          <XButton
            v-if="post.status !== 0"
            color="neutral"
            variant="ghost"
            size="sm"
            :disabled="actionsDisabled"
            @click="emit('setStatus', post, 0)"
          >
            恢复
          </XButton>
          <XButton
            v-if="post.status !== 1"
            color="neutral"
            variant="ghost"
            size="sm"
            :disabled="actionsDisabled"
            @click="emit('setStatus', post, 1)"
          >
            隐藏
          </XButton>
          <n-popconfirm
            v-if="post.status !== 2"
            positive-text="确认删除"
            negative-text="取消"
            @positive-click="emit('setStatus', post, 2)"
          >
            <template #trigger>
              <XButton
                color="danger"
                variant="ghost"
                size="sm"
                :disabled="actionsDisabled"
              >
                删除
              </XButton>
            </template>
            确定将帖子「{{ post.title }}」标记为已删除？
          </n-popconfirm>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.post-row,
.post-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.post-row {
  align-items: flex-start;
}

.post-content {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
}

.post-badges,
.post-metrics,
.action-group {
  display: flex;
  align-items: center;
  gap: 7px;
}

.post-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 4px 16px;
  margin-top: 4px;
  margin-left: -10px;
}

.action-group {
  flex-wrap: wrap;
  gap: 4px;
}

.status-actions {
  margin-left: auto;
}

.action-label {
  color: var(--color-muted);
  font-size: 12px;
  white-space: nowrap;
}

.pin-settings {
  display: flex;
  width: 220px;
  flex-direction: column;
  gap: 10px;
}

.pin-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.post-badges {
  flex-wrap: wrap;
}

.post-title {
  font-size: 17px;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.post-footer {
  margin-top: 3px;
}

.post-metrics span {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--color-muted);
  font-size: 12px;
}

@media (max-width: 680px) {
  .post-footer {
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }
}
</style>
