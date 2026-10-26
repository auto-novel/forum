<script setup lang="ts">
import { MoreVertOutlined } from '@vicons/material';
import { computed, onBeforeUnmount, ref } from 'vue';

import type { PostComment } from '@/api';
import {
  Toast,
  XDropdownMenu,
  XDropdownMenuItem,
  XDropdownMenuSeparator,
  XButton,
  XConfirmDialog,
  copyToClipboard,
  useWebKit,
} from '@novelia/web-kit';
import UserBlacklistDialog from '@/components/UserBlacklistDialog.vue';
import UserModerationDialog from '@/components/UserModerationDialog.vue';
import { useCommentStore } from '@/stores/comment';

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
const blacklistOpen = ref(false);
const canBlockAuthor = computed(
  () => whoami.value.isSignedIn && !isOwner.value,
);
const commentStore = useCommentStore();
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
    error: '复制失败，请手动选择评论内容复制',
  });
  copying.value = false;
}

async function removeComment() {
  submitting.value = true;
  await Toast.promise(
    commentStore
      .deleteComment(props.comment.id, asAdmin.value)
      .then(() => emit('statusChanged', 2)),
    {
      success: '评论已删除',
      error: '删除评论失败',
    },
  );
  submitting.value = false;
}

async function unhideComment() {
  if (submitting.value || !asAdmin.value || props.comment.status !== 1) return;
  submitting.value = true;
  await Toast.promise(
    commentStore
      .unhideComment(props.comment.id)
      .then(() => emit('statusChanged', 0)),
    {
      success: '评论已取消隐藏',
      error: '取消隐藏失败',
    },
  );
  submitting.value = false;
}

async function hideComment() {
  submitting.value = true;
  await Toast.promise(
    commentStore
      .hideComment(props.comment.id)
      .then(() => emit('statusChanged', 1)),
    {
      success: '评论已隐藏',
      error: '隐藏评论失败',
    },
  );
  submitting.value = false;
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
</script>

<template>
  <div class="ml-auto flex items-center gap-1">
    <XButton
      v-if="canReply"
      color="neutral"
      variant="ghost"
      size="xs"
      :aria-expanded="replying"
      @click="emit('reply')"
    >
      回复
    </XButton>
    <XButton
      v-if="canEdit"
      color="neutral"
      variant="ghost"
      size="xs"
      @click="emit('edit')"
    >
      编辑
    </XButton>
    <XDropdownMenu align="end" class="min-w-32 w-32">
      <template #trigger>
        <XButton
          color="neutral"
          variant="ghost"
          size="icon-xs"
          aria-label="更多操作"
          title="更多操作"
        >
          <MoreVertOutlined class="size-4" aria-hidden="true" />
        </XButton>
      </template>
      <XDropdownMenuItem :disabled="copying" @select="copyComment">
        复制原文
      </XDropdownMenuItem>
      <XDropdownMenuSeparator v-if="asAdmin || canDelete" />
      <XDropdownMenuItem
        v-if="asAdmin && isPublished"
        :disabled="submitting"
        @select="confirmationAction = 'hide'"
      >
        隐藏评论
      </XDropdownMenuItem>
      <XDropdownMenuItem
        v-if="asAdmin && comment.status === 1"
        :disabled="submitting"
        @select="unhideComment"
      >
        取消隐藏
      </XDropdownMenuItem>
      <XDropdownMenuItem
        v-if="canDelete"
        destructive
        :disabled="submitting"
        @select="confirmationAction = 'delete'"
      >
        删除评论
      </XDropdownMenuItem>
      <XDropdownMenuSeparator v-if="canBlockAuthor || canModerateAuthor" />
      <XDropdownMenuItem
        v-if="canBlockAuthor"
        :destructive="!comment.authorBlocked"
        :disabled="submitting"
        @select="blacklistOpen = true"
      >
        {{ comment.authorBlocked ? '取消屏蔽' : '屏蔽用户' }}
      </XDropdownMenuItem>
      <template v-if="canModerateAuthor">
        <XDropdownMenuItem
          destructive
          :disabled="submitting"
          @select="userModerationAction = 'strike'"
        >
          处罚用户
        </XDropdownMenuItem>
        <XDropdownMenuItem
          destructive
          :disabled="submitting"
          @select="userModerationAction = 'ban'"
        >
          封禁用户
        </XDropdownMenuItem>
      </template>
    </XDropdownMenu>

    <XConfirmDialog
      :open="confirmationAction != null"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :loading="submitting"
      danger
      @update:open="handleConfirmationOpenChange"
      @confirm="confirmAction"
    />
    <UserBlacklistDialog
      v-if="blacklistOpen && canBlockAuthor"
      :user-id="comment.authorId"
      :username="comment.authorUsername"
      :author-blocked="comment.authorBlocked"
      @close="blacklistOpen = false"
    />
    <UserModerationDialog
      v-if="userModerationAction"
      open
      :action="userModerationAction"
      :user-id="comment.authorId"
      :username="comment.authorUsername"
      :evidence="moderationEvidence"
      :comment-id="comment.id"
      @comment-deleted="emit('statusChanged', 2)"
      @update:open="handleUserModerationOpenChange"
      @comments-deleted="emit('authorCommentsDeleted')"
    />
  </div>
</template>
