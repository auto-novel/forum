import { ref } from 'vue';
import { defineStore } from 'pinia';

import {
  createPostComment,
  deletePostComment,
  getPostComments,
  setPostCommentStatus,
  updatePostComment,
  type PostComment,
} from '@/api';

interface CommentPageState {
  postId: number;
  page: number;
  pageSize: number;
  ids: number[];
  total: number;
  loading: boolean;
  error: string;
}

const EMPTY_PAGE_STATE: Readonly<CommentPageState> = Object.freeze({
  postId: 0,
  page: 1,
  pageSize: 1,
  ids: [],
  total: 0,
  loading: false,
  error: '',
});

function pageKey(postId: number, page: number, pageSize: number) {
  return `${postId}:${page}:${pageSize}`;
}

export const useCommentStore = defineStore('comment', () => {
  const commentsById = ref<Record<number, PostComment>>({});
  const pages = ref<Record<string, CommentPageState>>({});
  const pageControllers = new Map<string, AbortController>();

  function getPageState(postId: number, page: number, pageSize: number) {
    return pages.value[pageKey(postId, page, pageSize)] ?? EMPTY_PAGE_STATE;
  }

  function getPageComments(postId: number, page: number, pageSize: number) {
    return getPageState(postId, page, pageSize).ids.flatMap((id) => {
      const comment = commentsById.value[id];
      return comment ? [comment] : [];
    });
  }

  function ensurePageState(postId: number, page: number, pageSize: number) {
    const key = pageKey(postId, page, pageSize);
    return (pages.value[key] ??= {
      postId,
      page,
      pageSize,
      ids: [],
      total: 0,
      loading: false,
      error: '',
    });
  }

  function setComment(comment: PostComment) {
    commentsById.value[comment.id] = comment;
  }

  async function loadPage(postId: number, page: number, pageSize: number) {
    const key = pageKey(postId, page, pageSize);
    pageControllers.get(key)?.abort();
    const state = ensurePageState(postId, page, pageSize);
    state.ids = [];
    state.total = 0;
    state.error = '';

    if (!postId) {
      state.loading = false;
      return;
    }

    const controller = new AbortController();
    pageControllers.set(key, controller);
    state.loading = true;
    try {
      const result = await getPostComments(
        postId,
        { page, pageSize },
        controller.signal,
      );
      for (const comment of result.items) setComment(comment);
      state.ids = result.items.map((comment) => comment.id);
      state.total = result.total;
    } catch (reason) {
      if (reason instanceof DOMException && reason.name === 'AbortError')
        return;
      state.error = reason instanceof Error ? reason.message : '无法加载评论';
    } finally {
      if (pageControllers.get(key) === controller) {
        state.loading = false;
        pageControllers.delete(key);
      }
    }
  }

  async function createComment(
    postId: number,
    input: { content: string; rootId?: number },
  ) {
    const comment = await createPostComment(postId, input);
    setComment(comment);
    return comment;
  }

  function registerCreatedComment(
    comment: PostComment,
    currentPage: number,
    pageSize: number,
  ) {
    const key = pageKey(comment.postId, currentPage, pageSize);
    pageControllers.get(key)?.abort();
    pageControllers.delete(key);

    const currentState = ensurePageState(comment.postId, currentPage, pageSize);
    currentState.loading = false;
    currentState.error = '';

    const states = Object.values(pages.value).filter(
      (state) => state.postId === comment.postId,
    );
    for (const state of states) state.total += 1;

    const lastPage = Math.max(1, Math.ceil(currentState.total / pageSize));
    if (comment.rootId != null) {
      const lastReplyIndex = currentState.ids.reduce((result, id, index) => {
        const item = commentsById.value[id];
        return item &&
          (item.id === comment.rootId || item.rootId === comment.rootId)
          ? index
          : result;
      }, -1);
      currentState.ids.splice(lastReplyIndex + 1, 0, comment.id);
    } else if (currentPage === lastPage) {
      currentState.ids.push(comment.id);
    }
    return lastPage;
  }

  async function updateComment(id: number, content: string) {
    const comment = await updatePostComment(id, content);
    setComment(comment);
    return comment;
  }

  function applyStatus(id: number, status: number) {
    const comment = commentsById.value[id];
    if (!comment) return;
    setComment({ ...comment, status, content: '' });
  }

  async function deleteComment(id: number, asAdmin: boolean) {
    if (asAdmin) await setPostCommentStatus(id, 'deleted');
    else await deletePostComment(id);
    applyStatus(id, 2);
  }

  async function hideComment(id: number) {
    await setPostCommentStatus(id, 'hidden');
    applyStatus(id, 1);
  }

  return {
    getPageComments,
    getPageState,
    createComment,
    deleteComment,
    hideComment,
    loadPage,
    registerCreatedComment,
    updateComment,
  };
});
