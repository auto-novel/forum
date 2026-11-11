<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui';
import { MoreVertOutlined } from '@vicons/material';
import { computed, nextTick, onBeforeUnmount, ref } from 'vue';

import type { PostComment } from '@/api';
import { useForumApi } from '@/api';
import {
  Toast,
  UserBlacklistDialog,
  UserModerationDialog,
  XConfirmDialog,
  copyToClipboard,
  getApiErrorMessage,
  useWebKit,
} from '@novelia/web-kit';
import { useCommentStore } from '@/stores/comment';
import { usePostStore } from '@/stores/post';

const props = defineProps<{
  comment: PostComment;
  locked: boolean;
  postId: number;
  replying: boolean;
}>();

const emit = defineEmits<{
  reply: [];
  edit: [];
  statusChanged: [status: number];
  authorCommentsDeleted: [];
}>();

const { whoami } = useWebKit();
const forumApi = useForumApi();
const blacklistOpen = ref(false);
const menuTrigger = ref<{ $el: HTMLElement }>();
const canBlockAuthor = computed(
  () => whoami.value.isSignedIn && !isOwner.value,
);
const commentStore = useCommentStore();
const postStore = usePostStore();
const submitting = ref(false);
const copying = ref(false);
const now = ref(Date.now());
const confirmationAction = ref<'delete' | 'hide'>();
const userModerationAction = ref<'strike' | 'ban'>();

const modificationDeadline =
  new Date(props.comment.createdAt).getTime() + 30 * 60_000;
const expiryTimer = window.setTimeout(
  () => {
    now.value = Date.now();
  },
  Math.max(0, modificationDeadline - Date.now() + 50),
);

const asAdmin = computed(() => whoami.value.asAdmin);
const isOwner = computed(
  () => whoami.value.user?.id === props.comment.authorId,
);
const isPublished = computed(() => props.comment.status === 0);
const canReply = computed(
  () => isPublished.value && whoami.value.isSignedIn && !props.locked,
);
const canEdit = computed(
  () =>
    isPublished.value &&
    (asAdmin.value || (isOwner.value && now.value <= modificationDeadline)),
);
const canDelete = computed(() => asAdmin.value || canEdit.value);
const canModerateAuthor = computed(() => asAdmin.value && !isOwner.value);
const moderationEvidence = computed(() =>
  [
    `论坛评论 #${props.comment.id}（帖子 #${props.postId}）`,
    new URL(
      `/p/${props.postId}#comment-${props.comment.id}`,
      window.location.origin,
    ).toString(),
    '',
    '评论原文：',
    props.comment.content,
  ].join('\n'),
);
const confirmation = computed(() =>
  confirmationAction.value === 'delete'
    ? {
        title: '删除评论',
        description: '确定删除这条评论吗？',
        confirmLabel: '删除评论',
      }
    : {
        title: '隐藏评论',
        description: '确定隐藏这条评论吗？',
        confirmLabel: '隐藏评论',
      },
);

onBeforeUnmount(() => window.clearTimeout(expiryTimer));

async function copyComment() {
  if (copying.value) return;
  copying.value = true;
  await Toast.promise(copyToClipboard(props.comment.content), {
    success: '评论原文已复制',
    error: (reason) =>
      getApiErrorMessage(reason, '复制失败，请手动选择评论内容复制'),
  }).catch(() => {});
  copying.value = false;
}

async function removeComment() {
  submitting.value = true;
  await Toast.promise(
    commentStore.deleteComment(props.comment.id, asAdmin.value),
    {
      success: '评论已删除',
      error: (reason) => getApiErrorMessage(reason, '删除评论失败'),
    },
  )
    .then(() => emit('statusChanged', 2))
    .catch(() => {});
  submitting.value = false;
}

async function unhideComment() {
  if (submitting.value || !asAdmin.value || props.comment.status !== 1) return;
  submitting.value = true;
  await Toast.promise(commentStore.unhideComment(props.comment.id), {
    success: '评论已取消隐藏',
    error: (reason) => getApiErrorMessage(reason, '取消隐藏失败'),
  })
    .then(() => emit('statusChanged', 0))
    .catch(() => {});
  submitting.value = false;
}

async function hideComment() {
  submitting.value = true;
  await Toast.promise(commentStore.hideComment(props.comment.id), {
    success: '评论已隐藏',
    error: (reason) => getApiErrorMessage(reason, '隐藏评论失败'),
  })
    .then(() => emit('statusChanged', 1))
    .catch(() => {});
  submitting.value = false;
}

/** 由共享黑名单弹窗调用，请求成功后同步帖子与评论缓存。 */
async function setUserBlocked(userId: number, blocked: boolean) {
  await forumApi.setUserBlocked(userId, blocked);
  postStore.refreshBlacklist(userId, blocked);
  commentStore.refreshBlacklist(userId, blocked);
}

/** 由共享处罚弹窗调用，处罚后删除目标评论。 */
function deletePunishedComment() {
  return commentStore.deleteComment(props.comment.id, true);
}

/** 由共享处罚弹窗调用，删除评论后同步评论缓存。 */
async function deleteCommentsByAuthor(userId: number) {
  await forumApi.deleteCommentsByAuthor(userId);
  commentStore.registerDeletedCommentsByAuthor(userId);
}

function confirmAction() {
  const action = confirmationAction.value;
  confirmationAction.value = undefined;
  if (action === 'delete') void removeComment();
  else if (action === 'hide') void hideComment();
}

function handleConfirmationOpenChange(open: boolean) {
  if (!open) confirmationAction.value = undefined;
}

function handleUserModerationOpenChange(open: boolean) {
  if (!open) userModerationAction.value = undefined;
}

// 菜单项并不是受控弹窗的 DialogTrigger，显式保留关闭后的回焦目标。
function restoreMenuFocus(event?: Event) {
  if (!menuTrigger.value?.$el.isConnected) return;
  event?.preventDefault();
  menuTrigger.value.$el.focus({ preventScroll: true });
}

function handleMenuCloseAutoFocus(event: Event) {
  if (
    confirmationAction.value ||
    blacklistOpen.value ||
    userModerationAction.value
  )
    event.preventDefault();
}

function closeBlacklist() {
  blacklistOpen.value = false;
  void nextTick(() => restoreMenuFocus());
}

const menuItems = computed<DropdownMenuItem[][]>(() => {
  const management: DropdownMenuItem[] = [];
  if (asAdmin.value && isPublished.value)
    management.push({
      label: '隐藏评论',
      disabled: submitting.value,
      onSelect: () => {
        confirmationAction.value = 'hide';
      },
    });
  if (asAdmin.value && props.comment.status === 1)
    management.push({
      label: '取消隐藏',
      disabled: submitting.value,
      onSelect: unhideComment,
    });

  if (canDelete.value)
    management.push({
      label: '删除评论',
      color: 'error',
      disabled: submitting.value,
      onSelect: () => {
        confirmationAction.value = 'delete';
      },
    });
  const author: DropdownMenuItem[] = [];
  if (canBlockAuthor.value)
    author.push({
      label: props.comment.authorBlocked ? '取消屏蔽' : '屏蔽用户',
      color: props.comment.authorBlocked ? undefined : 'error',
      disabled: submitting.value,
      onSelect: () => {
        blacklistOpen.value = true;
      },
    });
  if (canModerateAuthor.value)
    author.push(
      {
        label: '处罚用户',
        color: 'error',
        disabled: submitting.value,
        onSelect: () => {
          userModerationAction.value = 'strike';
        },
      },
      {
        label: '封禁用户',
        color: 'error',
        disabled: submitting.value,
        onSelect: () => {
          userModerationAction.value = 'ban';
        },
      },
    );
  return [
    [{ label: '复制原文', disabled: copying.value, onSelect: copyComment }],
    ...(management.length ? [management] : []),
    ...(author.length ? [author] : []),
  ];
});
</script>

<template>
  <div class="ml-auto flex items-center gap-1">
    <UButton
      v-if="canReply"
      color="neutral"
      variant="ghost"
      size="sm"
      :aria-expanded="replying"
      @click="emit('reply')"
    >
      回复
    </UButton>
    <UButton
      v-if="canEdit"
      color="neutral"
      variant="ghost"
      size="sm"
      @click="emit('edit')"
    >
      编辑
    </UButton>
    <UDropdownMenu
      :items="menuItems"
      :content="{ align: 'end', onCloseAutoFocus: handleMenuCloseAutoFocus }"
      :ui="{ content: 'min-w-32 w-32' }"
    >
      <UButton
        ref="menuTrigger"
        color="neutral"
        variant="ghost"
        square
        size="xs"
        aria-label="更多操作"
        title="更多操作"
      >
        <MoreVertOutlined class="size-4" aria-hidden="true" />
      </UButton>
    </UDropdownMenu>

    <XConfirmDialog
      :open="confirmationAction != null"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :loading="submitting"
      danger
      @update:open="handleConfirmationOpenChange"
      @after:leave="restoreMenuFocus()"
      @confirm="confirmAction"
    />
    <UserBlacklistDialog
      v-if="blacklistOpen && canBlockAuthor"
      :user-id="comment.authorId"
      :username="comment.authorUsername"
      :author-blocked="comment.authorBlocked"
      :set-user-blocked="setUserBlocked"
      @close="closeBlacklist"
    />
    <UserModerationDialog
      v-if="userModerationAction"
      open
      :action="userModerationAction"
      :content="{ onCloseAutoFocus: restoreMenuFocus }"
      :user-id="comment.authorId"
      :username="comment.authorUsername"
      :evidence="moderationEvidence"
      :delete-comment="deletePunishedComment"
      :delete-comments-by-author="deleteCommentsByAuthor"
      @comment-deleted="emit('statusChanged', 2)"
      @update:open="handleUserModerationOpenChange"
      @comments-deleted="emit('authorCommentsDeleted')"
    />
  </div>
</template>
