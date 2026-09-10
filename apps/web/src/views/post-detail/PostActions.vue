<script setup lang="ts">
import {
  ChatBubbleOutlineOutlined,
  StarBorderOutlined,
  StarFilled,
} from '@vicons/material';
import { computed, ref, useTemplateRef, watch } from 'vue';

import {
  authUser,
  deletePost,
  moderatePost,
  setPostFavorite,
  type Post,
} from '@/api';

const props = defineProps<{ post: Post }>();

const emit = defineEmits<{
  edit: [];
  deleted: [];
  updated: [post: Post];
  comment: [];
}>();

const menu = useTemplateRef<HTMLDetailsElement>('menu');
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

function closeMenu() {
  if (menu.value) menu.value.open = false;
}

function editPost() {
  closeMenu();
  emit('edit');
}

async function removePost() {
  closeMenu();
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

async function updateModeration(
  changes: Partial<Pick<Post, 'commentsLocked' | 'pinOrder' | 'status'>>,
) {
  closeMenu();
  actionLoading.value = true;
  actionError.value = '';
  const nextPost = { ...props.post, ...changes };
  try {
    await moderatePost(props.post.id, {
      status: nextPost.status,
      commentsLocked: nextPost.commentsLocked,
      pinOrder: nextPost.pinOrder ?? null,
    });
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
  void updateModeration({ status: 1 });
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

      <details v-if="canManage" ref="menu" class="relative ml-auto">
        <summary class="post-action cursor-pointer list-none">更多 ···</summary>
        <div
          class="absolute top-[calc(100%+0.4rem)] right-0 z-20 w-40 rounded-md border border-border bg-surface p-1 shadow-xl"
        >
          <button
            type="button"
            class="post-menu-item"
            :disabled="actionLoading"
            @click="editPost"
          >
            编辑帖子
          </button>
          <template v-if="isAdmin">
            <button
              type="button"
              class="post-menu-item"
              :disabled="actionLoading"
              @click="
                updateModeration({
                  pinOrder: post.pinOrder == null ? 0 : undefined,
                })
              "
            >
              {{ post.pinOrder == null ? '置顶帖子' : '取消置顶' }}
            </button>
            <button
              type="button"
              class="post-menu-item"
              :disabled="actionLoading"
              @click="
                updateModeration({ commentsLocked: !post.commentsLocked })
              "
            >
              {{ post.commentsLocked ? '开放评论' : '锁定评论' }}
            </button>
            <button
              type="button"
              class="post-menu-item"
              :disabled="actionLoading"
              @click="hidePost"
            >
              隐藏帖子
            </button>
          </template>
          <button
            type="button"
            class="post-menu-item text-red-600"
            :disabled="actionLoading"
            @click="removePost"
          >
            删除帖子
          </button>
        </div>
      </details>
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

.post-action:focus-visible,
.post-menu-item:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
}

.post-menu-item {
  display: block;
  width: 100%;
  border-radius: 0.25rem;
  padding: 0.5rem 0.7rem;
  font-size: 0.8125rem;
  text-align: left;
}

.post-menu-item:hover {
  background: var(--color-paper);
}

.post-menu-item:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

summary::-webkit-details-marker {
  display: none;
}
</style>
