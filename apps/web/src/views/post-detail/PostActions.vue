<script setup lang="ts">
import {
  ChatBubbleOutlineOutlined,
  EditOutlined,
  MoreVertOutlined,
  ShareOutlined,
} from '@vicons/material';
import { computed, onScopeDispose, ref, watch } from 'vue';
import {
  Toast,
  XDropdownMenu,
  XDropdownMenuItem,
  XDropdownMenuSeparator,
  XButton,
  XConfirmDialog,
  copyToClipboard,
  getApiErrorMessage,
  useWebKit,
} from '@novelia/web-kit';

import { useForumApi, type Post } from '@/api';
import PostFavoriteButton from './PostFavoriteButton.vue';
import PostShareDialog from './PostShareDialog.vue';
import UserBlacklistDialog from '@/components/UserBlacklistDialog.vue';
import UserModerationDialog from '@/components/UserModerationDialog.vue';

const props = defineProps<{
  post: Post;
  canFavorite: boolean;
  favoriteLoading: boolean;
}>();

const emit = defineEmits<{
  favorite: [];
  comments: [];
  edit: [];
  deleted: [];
  updated: [patch: Partial<Pick<Post, 'pinOrder' | 'commentsLocked'>>];
  authorCommentsDeleted: [];
}>();

const forumApi = useForumApi();
const { whoami } = useWebKit();
const blacklistOpen = ref(false);
const canBlockAuthor = computed(
  () => whoami.value.isSignedIn && !isOwner.value,
);
const actionLoading = ref(false);
const copying = ref(false);
const shareOpen = ref(false);
const confirmationAction = ref<'delete' | 'hide'>();
const userModerationAction = ref<'strike' | 'ban'>();
let context = 0;
watch(
  () => [props.post.id, whoami.value.user?.id],
  (next, previous) => {
    if (next.some((value, index) => value !== previous[index])) {
      context++;
      actionLoading.value = false;
      shareOpen.value = false;
      blacklistOpen.value = false;
      confirmationAction.value = undefined;
      userModerationAction.value = undefined;
    }
  },
  { flush: 'sync' },
);
onScopeDispose(() => context++);
const asAdmin = computed(() => whoami.value.asAdmin);
const isOwner = computed(() => whoami.value.user?.id === props.post.authorId);
const canManagePost = computed(() => isOwner.value || asAdmin.value);
const canDelete = computed(
  () =>
    asAdmin.value ||
    (isOwner.value &&
      Date.now() <= new Date(props.post.createdAt).getTime() + 20 * 60_000),
);
const canModerateAuthor = computed(() => asAdmin.value && !isOwner.value);
const moderationEvidence = computed(() =>
  [
    `论坛帖子 #${props.post.id}：${props.post.title}`,
    new URL(`/p/${props.post.id}`, window.location.origin).toString(),
  ].join('\n'),
);
const confirmation = computed(() =>
  confirmationAction.value === 'delete'
    ? {
        title: '删除帖子',
        description: '确定删除这篇帖子吗？',
        confirmLabel: '删除帖子',
      }
    : {
        title: '隐藏帖子',
        description: '确定隐藏这篇帖子吗？',
        confirmLabel: '隐藏帖子',
      },
);

async function copyPost() {
  if (copying.value) return;
  copying.value = true;
  await Toast.promise(copyToClipboard(props.post.content), {
    success: '帖子原文已复制',
    error: '复制失败，请手动选择帖子内容复制',
  });
  copying.value = false;
}

function editPost() {
  emit('edit');
}

function removePost() {
  void updateModeration(
    () => forumApi.deletePost(props.post.id),
    { status: 1 },
    '帖子已删除',
    '删除帖子失败',
  );
}

async function updateModeration(
  request: () => Promise<unknown>,
  patch: Partial<Pick<Post, 'pinOrder' | 'commentsLocked' | 'status'>>,
  successMessage: string,
  failureMessage: string,
) {
  if (actionLoading.value || !canManagePost.value) return;
  const requestContext = context;
  const isCurrent = () => requestContext === context;
  actionLoading.value = true;
  try {
    // 保留手动提示，避免切换帖子、账号或离开页面后展示过期结果。
    await request();
    if (!isCurrent()) return;
    Toast.success(successMessage);
    if (patch.status != null && patch.status !== 0) emit('deleted');
    else emit('updated', patch);
  } catch (reason) {
    if (!isCurrent()) return;
    const message = await getApiErrorMessage(reason, failureMessage);
    if (isCurrent()) Toast.error(message);
  } finally {
    if (isCurrent()) actionLoading.value = false;
  }
}

function hidePost() {
  void updateModeration(
    () => forumApi.setPostStatus(props.post.id, 1),
    { status: 1 },
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

function handleUserModerationOpenChange(open: boolean) {
  if (!open) userModerationAction.value = undefined;
}

function togglePin() {
  const pinOrder = props.post.pinOrder == null ? 0 : null;
  const request = () =>
    pinOrder == null
      ? forumApi.unpinPost(props.post.id)
      : forumApi.pinPost(props.post.id, pinOrder);
  void updateModeration(
    request,
    { pinOrder },
    pinOrder == null ? '已取消置顶' : '帖子已置顶',
    pinOrder == null ? '取消置顶失败' : '置顶帖子失败',
  );
}

function toggleLock() {
  const commentsLocked = !props.post.commentsLocked;
  const request = () =>
    commentsLocked
      ? forumApi.lockPost(props.post.id)
      : forumApi.unlockPost(props.post.id);
  void updateModeration(
    request,
    { commentsLocked },
    commentsLocked ? '评论区已锁定' : '评论区已开放',
    commentsLocked ? '锁定评论失败' : '开放评论失败',
  );
}
</script>

<template>
  <section class="relative" aria-label="帖子快捷操作">
    <div class="flex flex-wrap items-center gap-2">
      <PostFavoriteButton
        v-if="canFavorite"
        prominent
        :favorited="post.favorited"
        :loading="favoriteLoading"
        @toggle="emit('favorite')"
      />
      <XButton
        as="a"
        href="#comments"
        variant="outline"
        size="sm"
        class="text-ink!"
        aria-controls="comments"
        @click.prevent="emit('comments')"
      >
        <ChatBubbleOutlineOutlined class="size-4" aria-hidden="true" />
        评论
      </XButton>
      <XButton
        v-if="canManagePost"
        variant="outline"
        size="sm"
        class="text-ink!"
        :disabled="actionLoading"
        @click="editPost"
      >
        <EditOutlined class="size-4" aria-hidden="true" />
        编辑
      </XButton>
      <XButton
        variant="ghost"
        size="icon"
        class="w-auto min-h-9 px-1.5 text-ink!"
        aria-label="分享帖子"
        title="分享帖子"
        aria-haspopup="dialog"
        :aria-expanded="shareOpen"
        @click="shareOpen = true"
      >
        <ShareOutlined class="size-5" aria-hidden="true" />
      </XButton>
      <span class="inline-flex" :inert="actionLoading">
        <XDropdownMenu align="end" class="w-40">
          <template #trigger>
            <XButton
              variant="ghost"
              size="icon"
              aria-label="更多操作"
              title="更多操作"
            >
              <MoreVertOutlined class="size-5" aria-hidden="true" />
            </XButton>
          </template>
          <XDropdownMenuItem :disabled="copying" @select="copyPost">
            复制原文
          </XDropdownMenuItem>
          <XDropdownMenuSeparator v-if="asAdmin || canDelete" />
          <template v-if="asAdmin">
            <XDropdownMenuItem :disabled="actionLoading" @select="togglePin">
              {{ post.pinOrder == null ? '置顶帖子' : '取消置顶' }}
            </XDropdownMenuItem>
            <XDropdownMenuItem :disabled="actionLoading" @select="toggleLock">
              {{ post.commentsLocked ? '开放评论' : '锁定评论' }}
            </XDropdownMenuItem>
            <XDropdownMenuItem
              :disabled="actionLoading"
              @select="confirmationAction = 'hide'"
            >
              隐藏帖子
            </XDropdownMenuItem>
          </template>
          <XDropdownMenuItem
            v-if="canDelete"
            destructive
            :disabled="actionLoading"
            @select="confirmationAction = 'delete'"
          >
            删除帖子
          </XDropdownMenuItem>
          <XDropdownMenuSeparator v-if="canBlockAuthor || canModerateAuthor" />
          <XDropdownMenuItem
            v-if="canBlockAuthor"
            :destructive="!post.authorBlocked"
            :disabled="actionLoading"
            @select="blacklistOpen = true"
          >
            {{ post.authorBlocked ? '取消屏蔽' : '屏蔽用户' }}
          </XDropdownMenuItem>
          <template v-if="canModerateAuthor">
            <XDropdownMenuItem
              destructive
              :disabled="actionLoading"
              @select="userModerationAction = 'strike'"
            >
              处罚用户
            </XDropdownMenuItem>
            <XDropdownMenuItem
              destructive
              :disabled="actionLoading"
              @select="userModerationAction = 'ban'"
            >
              封禁用户
            </XDropdownMenuItem>
          </template>
        </XDropdownMenu>
      </span>
    </div>
    <PostShareDialog
      :open="shareOpen"
      :post="post"
      @update:open="shareOpen = $event"
    />
    <XConfirmDialog
      :open="confirmationAction != null"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :loading="actionLoading"
      danger
      @update:open="handleConfirmationOpenChange"
      @confirm="confirmAction"
    />
    <UserBlacklistDialog
      v-if="blacklistOpen && canBlockAuthor"
      :user-id="post.authorId"
      :username="post.authorUsername"
      :author-blocked="post.authorBlocked"
      @close="blacklistOpen = false"
    />
    <UserModerationDialog
      v-if="userModerationAction"
      open
      :action="userModerationAction"
      :user-id="post.authorId"
      :username="post.authorUsername"
      :evidence="moderationEvidence"
      @update:open="handleUserModerationOpenChange"
      @comments-deleted="emit('authorCommentsDeleted')"
    />
  </section>
</template>
