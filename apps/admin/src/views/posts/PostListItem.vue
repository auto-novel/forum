<script setup lang="ts">
import {
  CommentOutlined,
  LockOutlined,
  PushPinOutlined,
  VisibilityOutlined,
} from '@vicons/material';
import {
  NButton,
  NIcon,
  NInputNumber,
  NPopconfirm,
  NSpin,
  NTag,
  NText,
} from 'naive-ui';
import { ref, watch } from 'vue';

import type { PostSummary } from '@/api';

const props = defineProps<{
  post: PostSummary;
  categoryName: string;
  actionsDisabled: boolean;
  saving: boolean;
}>();

const emit = defineEmits<{
  setStatus: [post: PostSummary, status: number];
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
      <div class="post-badges">
        <n-tag size="small" :bordered="false">{{ categoryName }}</n-tag>
        <n-tag
          v-if="post.status === 1"
          size="small"
          type="warning"
          :bordered="false"
        >
          隐藏
        </n-tag>
        <n-tag
          v-if="post.status === 2"
          size="small"
          type="error"
          :bordered="false"
        >
          已删除
        </n-tag>
        <n-tag
          v-if="post.pinOrder != null"
          size="small"
          type="warning"
          :bordered="false"
        >
          <template #icon><n-icon :component="PushPinOutlined" /></template>
          置顶 {{ post.pinOrder }}
        </n-tag>
        <n-tag
          v-if="post.commentsLocked"
          size="small"
          type="error"
          :bordered="false"
        >
          <template #icon><n-icon :component="LockOutlined" /></template>
          已锁评
        </n-tag>
      </div>
      <n-text strong class="post-title">{{ post.title }}</n-text>
      <div class="post-footer">
        <n-text depth="3">
          {{ post.authorUsername }} (ID {{ post.authorId }}) ·
          {{ formatDate(post.activeAt) }}
        </n-text>
        <div class="post-metrics">
          <span>
            <n-icon :component="VisibilityOutlined" />
            {{ post.viewsCount }}
          </span>
          <span>
            <n-icon :component="CommentOutlined" />
            {{ post.commentsCount }}
          </span>
          <n-button
            text
            type="primary"
            size="small"
            @click="emit('reviewComments', post)"
          >
            审核评论
          </n-button>
        </div>
      </div>
      <div class="post-actions">
        <n-spin v-if="saving" size="small" />
        <span class="action-label">状态</span>
        <n-button
          size="small"
          :type="post.status === 0 ? 'primary' : 'default'"
          :disabled="actionsDisabled || post.status === 0"
          @click="emit('setStatus', post, 0)"
        >
          正常发布
        </n-button>
        <n-button
          size="small"
          :type="post.status === 1 ? 'warning' : 'default'"
          :disabled="actionsDisabled || post.status === 1"
          @click="emit('setStatus', post, 1)"
        >
          隐藏
        </n-button>
        <n-popconfirm
          v-if="post.status !== 2"
          positive-text="确认删除"
          negative-text="取消"
          @positive-click="emit('setStatus', post, 2)"
        >
          <template #trigger>
            <n-button size="small" type="error" :disabled="actionsDisabled">
              删除
            </n-button>
          </template>
          确定将帖子「{{ post.title }}」标记为已删除？
        </n-popconfirm>
        <n-button v-else size="small" type="error" disabled>已删除</n-button>
        <span class="action-divider" />
        <n-button
          size="small"
          :disabled="actionsDisabled"
          @click="emit('setCommentsLocked', post, !post.commentsLocked)"
        >
          {{ post.commentsLocked ? '解锁评论' : '锁定评论' }}
        </n-button>
        <span class="action-divider" />
        <span class="action-label">置顶顺序</span>
        <n-input-number
          v-model:value="pinOrderInput"
          class="pin-input"
          size="small"
          :min="-2147483648"
          :max="2147483647"
          :precision="0"
          :show-button="false"
          :disabled="actionsDisabled"
          placeholder="顺序"
        />
        <n-button
          size="small"
          :disabled="
            actionsDisabled ||
            pinOrderInput == null ||
            pinOrderInput === (post.pinOrder ?? null)
          "
          @click="emit('setPinOrder', post, pinOrderInput)"
        >
          {{ post.pinOrder == null ? '置顶' : '更新' }}
        </n-button>
        <n-button
          v-if="post.pinOrder != null"
          size="small"
          :disabled="actionsDisabled"
          @click="emit('setPinOrder', post, null)"
        >
          取消置顶
        </n-button>
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
.post-actions {
  display: flex;
  align-items: center;
  gap: 7px;
}

.post-actions {
  flex-wrap: wrap;
  margin-top: 8px;
}

.action-label {
  color: var(--n-text-color-3);
  font-size: 12px;
  white-space: nowrap;
}

.action-divider {
  width: 1px;
  height: 18px;
  margin-inline: 5px;
  background: var(--n-border-color);
}

.pin-input {
  width: 96px;
}

.post-title {
  font-size: 17px;
  line-height: 1.4;
}

.post-footer {
  margin-top: 3px;
}

.post-metrics span {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--n-text-color-3);
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
