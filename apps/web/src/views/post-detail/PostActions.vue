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
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { notifyError, notifySuccess } from '@/notifications';
import { getApiErrorMessage } from '@/utils/apiError';

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
const confirmationAction = ref<'delete' | 'hide'>();
const postActionClass =
  'inline-flex min-h-9 items-center gap-[0.4rem] rounded-sm px-[0.7rem] text-[0.8125rem] font-semibold transition-colors duration-150 hover:bg-paper hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary disabled:cursor-not-allowed disabled:opacity-50';

const isAdmin = computed(() => authUser.value?.role === 'admin');
const canManage = computed(
  () =>
    authUser.value?.id === props.post.authorId ||
    authUser.value?.role === 'admin',
);
const confirmation = computed(() =>
  confirmationAction.value === 'delete'
    ? {
        title: '删除帖子',
        description: '确定删除这篇帖子吗？删除后无法恢复。',
        confirmLabel: '删除帖子',
      }
    : {
        title: '隐藏帖子',
        description: '确定隐藏这篇帖子吗？隐藏后可在管理端恢复。',
        confirmLabel: '隐藏帖子',
      },
);

async function toggleFavorite() {
  if (!authUser.value || favoriteLoading.value) return;
  favoriteLoading.value = true;
  const nextValue = !favorited.value;
  try {
    await setPostFavorite(props.post.id, nextValue);
    favorited.value = nextValue;
    emit('updated', { ...props.post, favorited: nextValue });
    notifySuccess(nextValue ? '帖子已收藏' : '已取消收藏');
  } catch (reason) {
    notifyError(await getApiErrorMessage(reason, '更新收藏失败'));
  } finally {
    favoriteLoading.value = false;
  }
}

function editPost() {
  emit('edit');
}

async function removePost() {
  actionLoading.value = true;
  try {
    await deletePost(props.post.id);
    notifySuccess('帖子已删除');
    emit('deleted');
  } catch (reason) {
    notifyError(await getApiErrorMessage(reason, '删除帖子失败'));
  } finally {
    actionLoading.value = false;
  }
}

async function updateModeration(
  request: Promise<unknown>,
  nextPost: Post,
  successMessage: string,
  failureMessage: string,
) {
  actionLoading.value = true;
  try {
    await request;
    notifySuccess(successMessage);
    if (nextPost.status !== 0) emit('deleted');
    else emit('updated', nextPost);
  } catch (reason) {
    notifyError(await getApiErrorMessage(reason, failureMessage));
  } finally {
    actionLoading.value = false;
  }
}

function hidePost() {
  void updateModeration(
    setPostStatus(props.post.id, 1),
    {
      ...props.post,
      status: 1,
    },
    '帖子已隐藏',
    '隐藏帖子失败',
  );
}

function confirmAction() {
  const action = confirmationAction.value;
  confirmationAction.value = undefined;
  if (action === 'delete') void removePost();
  else if (action === 'hide') hidePost();
}

function handleConfirmationOpenChange(open: boolean) {
  if (!open) confirmationAction.value = undefined;
}

function togglePin() {
  const pinOrder = props.post.pinOrder == null ? 0 : undefined;
  const request =
    pinOrder == null
      ? unpinPost(props.post.id)
      : pinPost(props.post.id, pinOrder);
  void updateModeration(
    request,
    { ...props.post, pinOrder },
    pinOrder == null ? '已取消置顶' : '帖子已置顶',
    pinOrder == null ? '取消置顶失败' : '置顶帖子失败',
  );
}

function toggleLock() {
  const commentsLocked = !props.post.commentsLocked;
  const request = commentsLocked
    ? lockPost(props.post.id)
    : unlockPost(props.post.id);
  void updateModeration(
    request,
    { ...props.post, commentsLocked },
    commentsLocked ? '评论区已锁定' : '评论区已开放',
    commentsLocked ? '锁定评论失败' : '开放评论失败',
  );
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
        :class="[postActionClass, favorited ? 'text-primary' : 'text-muted']"
        :disabled="favoriteLoading"
        :aria-pressed="favorited"
        @click="toggleFavorite"
      >
        <StarFilled v-if="favorited" class="size-4" aria-hidden="true" />
        <StarBorderOutlined v-else class="size-4" aria-hidden="true" />
        {{ favoriteLoading ? '处理中…' : favorited ? '取消收藏' : '收藏' }}
      </button>

      <button
        type="button"
        :class="[postActionClass, 'text-muted']"
        @click="emit('comment')"
      >
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
            <ActionMenuItem
              :disabled="actionLoading"
              @activate="confirmationAction = 'hide'"
            >
              隐藏帖子
            </ActionMenuItem>
          </template>
          <ActionMenuItem
            danger
            :disabled="actionLoading"
            @activate="confirmationAction = 'delete'"
          >
            删除帖子
          </ActionMenuItem>
        </ActionMenu>
      </div>
    </div>
    <ConfirmDialog
      :open="confirmationAction != null"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :loading="actionLoading"
      danger
      @update:open="handleConfirmationOpenChange"
      @confirm="confirmAction"
    />
  </section>
</template>
