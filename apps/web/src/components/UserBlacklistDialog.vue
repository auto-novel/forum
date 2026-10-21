<script setup lang="ts">
import { computed, onScopeDispose, ref, watch } from 'vue';
import {
  getApiErrorMessage,
  Toast,
  useWebKit,
  XConfirmDialog,
} from '@novelia/web-kit';
import { useForumApi } from '@/api';
import { usePostStore } from '@/stores/post';
import { useCommentStore } from '@/stores/comment';

const props = defineProps<{
  userId: number;
  username: string;
  authorBlocked: boolean;
}>();
const emit = defineEmits<{ close: [] }>();
const forumApi = useForumApi();
const { whoami } = useWebKit();
const posts = usePostStore();
const comments = useCommentStore();
const pending = ref(false);
let context = 0;
let active = true;
onScopeDispose(() => {
  active = false;
});
// Keep the requested operation stable while the content cache updates.
const blocked = props.authorBlocked;
const title = computed(
  () => `${blocked ? '取消屏蔽' : '屏蔽'} @${props.username}`,
);
watch(
  () => [props.userId, whoami.value.user?.id],
  () => {
    context++;
    emit('close');
  },
  { flush: 'sync' },
);

async function confirm() {
  const viewerId = whoami.value.user?.id;
  const userId = props.userId;
  if (pending.value || !viewerId || viewerId === userId) return;
  const requestContext = context;
  const isCurrent = () =>
    requestContext === context && whoami.value.user?.id === viewerId;
  pending.value = true;
  try {
    await forumApi.setUserBlocked(userId, !blocked);
    if (!isCurrent()) return;
    // Keep content caches in sync even if navigation closed this dialog.
    posts.refreshBlacklist(userId, !blocked);
    comments.refreshBlacklist(userId, !blocked);
    if (active) {
      Toast.success(blocked ? '已取消屏蔽' : '已屏蔽用户');
      emit('close');
    }
  } catch (reason) {
    if (!active || !isCurrent()) return;
    const message = await getApiErrorMessage(
      reason,
      blocked ? '取消屏蔽失败' : '屏蔽失败',
    );
    if (active && isCurrent()) Toast.error(message);
  } finally {
    pending.value = false;
  }
}
</script>

<template>
  <XConfirmDialog
    open
    :title="title"
    :description="
      blocked
        ? '取消后，该用户的帖子和评论将重新出现在列表中。'
        : '屏蔽后，帖子和收藏列表将过滤该用户的帖子，评论区将过滤其评论及其发起的评论串，仅对你生效。可在账户设置页面管理你的黑名单。'
    "
    :confirm-label="blocked ? '取消屏蔽' : '屏蔽用户'"
    :loading="pending"
    :danger="!blocked"
    @update:open="
      (open) => {
        if (!open && !pending) emit('close');
      }
    "
    @confirm="confirm"
  />
</template>
