<script setup lang="ts">
import {
  ChatBubbleOutlineOutlined,
  StarBorderOutlined,
  StarFilled,
} from '@vicons/material';
import { computed, ref, watch } from 'vue';

import {
  authUser,
  deletePost,
  lockPost,
  pinPost,
  setPostFavorite,
  setPostStatus,
  type Post,
  unlockPost,
  unpinPost,
} from '@/api';
import ActionMenu from '@/components/ActionMenu.vue';
import ActionMenuItem from '@/components/ActionMenuItem.vue';

const props = defineProps<{ post: Post }>();

const emit = defineEmits<{
  edit: [];
  deleted: [];
  updated: [post: Post];
  comment: [];
}>();

const favorited = ref(props.post.favorited);
const favoriteLoading = ref(false);
const actionLoading = ref(false);
const actionError = ref('');

const isAdmin = computed(() => authUser.value?.role === 'admin');
const canManage = computed(
  () =>
    authUser.value?.id === props.post.authorId ||
    authUser.value?.role === 'admin',
);

async function errorMessage(reason: unknown, fallback: string) {
  if (reason && typeof reason === 'object' && 'response' in reason) {
    const response = (reason as { response?: Response }).response;
    if (response) {
      try {
        return (await response.text()) || fallback;
      } catch {
        // Use the fallback below.
      }
    }
  }
  return reason instanceof Error ? reason.message : fallback;
}

async function toggleFavorite() {
  if (!authUser.value || favoriteLoading.value) return;
  favoriteLoading.value = true;
  actionError.value = '';
  const nextValue = !favorited.value;
  try {
    await setPostFavorite(props.post.id, nextValue);
    favorited.value = nextValue;
    emit('updated', { ...props.post, favorited: nextValue });
  } catch (reason) {
    actionError.value = await errorMessage(reason, '更新收藏失败');
  } finally {
    favoriteLoading.value = false;
  }
}

function editPost() {
  emit('edit');
}

async function removePost() {
  if (!window.confirm('确定删除这篇帖子吗？删除后无法恢复。')) return;
  actionLoading.value = true;
  actionError.value = '';
  try {
    await deletePost(props.post.id);
    emit('deleted');
  } catch (reason) {
    actionError.value = await errorMessage(reason, '删除帖子失败');
  } finally {
    actionLoading.value = false;
  }
}

async function updateModeration(request: Promise<unknown>, nextPost: Post) {
  actionLoading.value = true;
  actionError.value = '';
  try {
    await request;
    if (nextPost.status !== 0) emit('deleted');
    else emit('updated', nextPost);
  } catch (reason) {
    actionError.value = await errorMessage(reason, '管理帖子失败');
  } finally {
    actionLoading.value = false;
  }
}

function hidePost() {
  if (!window.confirm('确定隐藏这篇帖子吗？隐藏后可在管理端恢复。')) return;
  void updateModeration(setPostStatus(props.post.id, 1), {
    ...props.post,
    status: 1,
  });
}

function togglePin() {
  const pinOrder = props.post.pinOrder == null ? 0 : undefined;
  const request =
    pinOrder == null
      ? unpinPost(props.post.id)
      : pinPost(props.post.id, pinOrder);
  void updateModeration(request, { ...props.post, pinOrder });
}

function toggleLock() {
  const commentsLocked = !props.post.commentsLocked;
  const request = commentsLocked
    ? lockPost(props.post.id)
    : unlockPost(props.post.id);
  void updateModeration(request, { ...props.post, commentsLocked });
}

watch(
  () => props.post.favorited,
  (value) => {
    favorited.value = value;
  },
);
</script>

<template>
  <section
    class="relative mt-px rounded-sm bg-surface px-4 py-3 sm:px-6"
    aria-label="帖子操作"
  >
    <div class="flex flex-wrap items-center gap-2">
      <button
        v-if="authUser"
        type="button"
        class="post-action"
        :class="favorited ? 'post-action--active' : ''"
        :disabled="favoriteLoading"
        :aria-pressed="favorited"
        @click="toggleFavorite"
      >
        <StarFilled v-if="favorited" class="size-4" aria-hidden="true" />
        <StarBorderOutlined v-else class="size-4" aria-hidden="true" />
        {{ favoriteLoading ? '处理中…' : favorited ? '取消收藏' : '收藏' }}
      </button>

      <button type="button" class="post-action" @click="emit('comment')">
        <ChatBubbleOutlineOutlined class="size-4" aria-hidden="true" />
        评论
      </button>

      <div v-if="canManage" class="ml-auto">
        <ActionMenu>
          <ActionMenuItem :disabled="actionLoading" @activate="editPost">
            编辑帖子
          </ActionMenuItem>
          <template v-if="isAdmin">
            <ActionMenuItem :disabled="actionLoading" @activate="togglePin">
              {{ post.pinOrder == null ? '置顶帖子' : '取消置顶' }}
            </ActionMenuItem>
            <ActionMenuItem :disabled="actionLoading" @activate="toggleLock">
              {{ post.commentsLocked ? '开放评论' : '锁定评论' }}
            </ActionMenuItem>
            <ActionMenuItem :disabled="actionLoading" @activate="hidePost">
              隐藏帖子
            </ActionMenuItem>
          </template>
          <ActionMenuItem
            danger
            :disabled="actionLoading"
            @activate="removePost"
          >
            删除帖子
          </ActionMenuItem>
        </ActionMenu>
      </div>
    </div>
    <p v-if="actionError" class="mt-2 text-xs text-red-600" role="alert">
      {{ actionError }}
    </p>
  </section>
</template>

<style scoped>
.post-action {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  gap: 0.4rem;
  border-radius: 0.25rem;
  padding-inline: 0.7rem;
  color: var(--color-muted);
  font-size: 0.8125rem;
  font-weight: 600;
  transition:
    color 150ms ease,
    background-color 150ms ease;
}

.post-action:hover {
  background: var(--color-paper);
  color: var(--color-primary);
}

.post-action--active {
  color: var(--color-primary);
}

.post-action:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.post-action:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
}
</style>
