<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';

import type { PostComment } from '@/api';
import {
  COMMENT_REPLY_PAGE_SIZE,
  useCommentReplyPageQuery,
} from '@/stores/comment';
import { XAsyncContent, XPagination } from '@novelia/web-kit';

import CommentListItem from './CommentListItem.vue';

const REPLY_PAGE_SIZE = COMMENT_REPLY_PAGE_SIZE;

const props = defineProps<{
  comment: PostComment;
  locked: boolean;
  postId: number;
  replyToId?: number;
}>();

const emit = defineEmits<{
  reply: [comment: PostComment];
  cancelReply: [];
  created: [comment: PostComment];
  statusChanged: [id: number, status: number];
  authorCommentsDeleted: [];
}>();

const replyPage = ref(1);
const { replies, total, loading, error, refresh, retry } =
  useCommentReplyPageQuery(
    () => props.postId,
    () => props.comment.id,
    replyPage,
    REPLY_PAGE_SIZE,
    () => props.comment.replyCount > 0,
  );
const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / REPLY_PAGE_SIZE)),
);

async function handleCreated(comment: PostComment) {
  // 二级评论按时间正序排列，新回复位于最后一页。
  const lastPage = Math.max(
    1,
    Math.ceil((props.comment.replyCount + 1) / REPLY_PAGE_SIZE),
  );
  emit('created', comment);
  if (comment.rootId !== props.comment.id) return;
  replyPage.value = lastPage;
  await nextTick();
  await refresh();
  await nextTick();
  document
    .querySelector(`#comment-${comment.id}`)
    ?.scrollIntoView({ behavior: 'smooth', block: 'center' });
}
</script>

<template>
  <section>
    <CommentListItem
      :comment="comment"
      :locked="locked"
      :post-id="postId"
      :replying="replyToId === comment.id"
      @reply="emit('reply', $event)"
      @cancel-reply="emit('cancelReply')"
      @created="handleCreated"
      @status-changed="(id, status) => emit('statusChanged', id, status)"
      @author-comments-deleted="emit('authorCommentsDeleted')"
    />

    <div v-if="comment.replyCount > 0">
      <XAsyncContent
        :loading="loading"
        :error="error"
        size="compact"
        error-title=""
        retry-label="重试"
        :ui="{ error: 'min-h-0 pl-8' }"
        @retry="retry"
      >
        <template #loading>
          <div class="ml-8" role="status">
            <span class="sr-only">正在加载回复…</span>
            <div aria-hidden="true">
              <div
                v-for="index in Math.min(comment.replyCount, 3)"
                :key="index"
                class="py-4"
              >
                <div class="h-5 w-28 skeleton bg-border" />
                <div class="mt-0.5 rounded-floating bg-embedded px-4 py-3">
                  <div class="h-4 w-full skeleton" />
                  <div class="mt-2 h-4 w-2/3 skeleton" />
                </div>
              </div>
            </div>
          </div>
        </template>

        <CommentListItem
          v-for="reply in replies"
          :key="reply.id"
          :comment="reply"
          :locked="locked"
          :post-id="postId"
          :replying="replyToId === reply.id"
          @reply="emit('reply', $event)"
          @cancel-reply="emit('cancelReply')"
          @created="handleCreated"
          @status-changed="(id, status) => emit('statusChanged', id, status)"
          @author-comments-deleted="emit('authorCommentsDeleted')"
        />
        <XPagination
          v-if="totalPages > 1"
          :page="replyPage"
          :total-pages="totalPages"
          @update:page="replyPage = $event"
        />
      </XAsyncContent>
    </div>
  </section>
</template>
