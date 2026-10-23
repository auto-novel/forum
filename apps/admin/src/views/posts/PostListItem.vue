<script setup lang="ts">
import {
  CommentOutlined,
  LockOutlined,
  PushPinOutlined,
  VisibilityOutlined,
} from '@vicons/material';
import {
  XButton,
  XConfirmDialog,
  XDialog,
  XNumberInput,
} from '@novelia/web-kit';
import { ref, watch } from 'vue';

import AdminSpinner from '@/components/AdminSpinner.vue';
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

const deleteConfirmationOpen = ref(false);
const pinDialogOpen = ref(false);
const pinOrderInput = ref<number | null>(props.post.pinOrder ?? null);
watch(
  () => props.post.pinOrder,
  (value) => (pinOrderInput.value = value ?? null),
);

function openPinSettings() {
  pinOrderInput.value = props.post.pinOrder ?? null;
  pinDialogOpen.value = true;
}

function savePinOrder() {
  if (pinOrderInput.value == null) return;
  pinDialogOpen.value = false;
  emit('setPinOrder', props.post, pinOrderInput.value);
}

function clearPinOrder() {
  pinDialogOpen.value = false;
  emit('setPinOrder', props.post, null);
}

function confirmDelete() {
  if (
    !deleteConfirmationOpen.value ||
    props.actionsDisabled ||
    props.post.status === 2
  )
    return;
  deleteConfirmationOpen.value = false;
  emit('setStatus', props.post, 2);
}

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
      <div class="flex flex-col gap-1">
        <p class="post-title font-medium text-ink">{{ post.title }}</p>
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
          <XButton
            color="neutral"
            variant="ghost"
            size="sm"
            :disabled="actionsDisabled"
            @click="openPinSettings"
          >
            {{ post.pinOrder == null ? '置顶' : '置顶设置' }}
          </XButton>
          <XDialog
            v-model:open="pinDialogOpen"
            title="置顶设置"
            class="w-11/12 max-w-sm"
          >
            <div class="grid gap-2">
              <label :for="`pin-order-${post.id}`" class="text-sm">
                置顶顺序
              </label>
              <XNumberInput
                v-model="pinOrderInput"
                :id="`pin-order-${post.id}`"
                :min="-2147483648"
                :max="2147483647"
                :disabled="actionsDisabled"
                placeholder="输入顺序"
              />
              <p class="text-xs text-muted">数值越小越靠前。</p>
            </div>
            <template #footer>
              <div class="flex justify-end gap-2">
                <XButton
                  color="neutral"
                  variant="ghost"
                  size="sm"
                  @click="pinDialogOpen = false"
                >
                  取消
                </XButton>
                <XButton
                  v-if="post.pinOrder != null"
                  color="neutral"
                  variant="outline"
                  size="sm"
                  :disabled="actionsDisabled"
                  @click="clearPinOrder"
                >
                  取消置顶
                </XButton>
                <XButton
                  color="primary"
                  variant="soft"
                  size="sm"
                  :disabled="
                    actionsDisabled ||
                    pinOrderInput == null ||
                    pinOrderInput === (post.pinOrder ?? null)
                  "
                  @click="savePinOrder"
                >
                  保存
                </XButton>
              </div>
            </template>
          </XDialog>
        </div>
        <div class="action-group status-actions">
          <AdminSpinner v-if="saving" size="sm" label="正在保存" />
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
          <XButton
            v-if="post.status !== 2"
            color="danger"
            variant="ghost"
            size="sm"
            :disabled="actionsDisabled"
            @click="deleteConfirmationOpen = true"
          >
            删除
          </XButton>
          <XConfirmDialog
            v-if="post.status !== 2"
            v-model:open="deleteConfirmationOpen"
            title="确认删除帖子"
            :description="`确定将帖子「${post.title}」标记为已删除？`"
            confirm-label="确认删除"
            :loading="actionsDisabled"
            danger
            @confirm="confirmDelete"
          />
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

.post-badges {
  flex-wrap: wrap;
}

.post-title {
  font-size: 17px;
  line-height: 1.4;
  overflow-wrap: anywhere;
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
