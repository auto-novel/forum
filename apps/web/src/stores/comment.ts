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
  fetchedAt: number;
  accessedAt: number;
  revision: number;
  loading: boolean;
  error: string;
}

const CACHE_MAX_AGE = 60 * 1000;
const CACHE_MAX_PAGES = 30;
const EMPTY_PAGE_STATE: Readonly<CommentPageState> = Object.freeze({
  postId: 0,
  page: 1,
  pageSize: 1,
  ids: [],
  total: 0,
  fetchedAt: 0,
  accessedAt: 0,
  revision: 0,
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
  const pageRequests = new Map<string, Promise<void>>();
  const commentRevisions = new Map<number, number>();
  let mutationRevision = 0;

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
    if (!pages.value[key]) {
      pages.value[key] = {
        postId,
        page,
        pageSize,
        ids: [],
        total: 0,
        fetchedAt: 0,
        accessedAt: Date.now(),
        revision: 0,
        loading: false,
        error: '',
      };
    }
    return pages.value[key];
  }

  function setComment(comment: PostComment) {
    commentsById.value[comment.id] = comment;
  }

  function setMutatedComment(comment: PostComment) {
    mutationRevision += 1;
    commentRevisions.set(comment.id, mutationRevision);
    setComment(comment);
  }

  function prunePageCache(retainedKey: string) {
    const entries = Object.entries(pages.value);
    if (entries.length <= CACHE_MAX_PAGES) return;

    const evictedPages = entries
      .filter(
        ([key, state]) =>
          key !== retainedKey && !state.loading && !pageControllers.has(key),
      )
      .sort(([, left], [, right]) => left.accessedAt - right.accessedAt)
      .slice(0, entries.length - CACHE_MAX_PAGES);
    const evictedIds = evictedPages.flatMap(([, state]) => state.ids);
    for (const [key] of evictedPages) delete pages.value[key];

    const retainedIds = new Set(
      Object.values(pages.value).flatMap((state) => state.ids),
    );
    for (const id of evictedIds) {
      if (!retainedIds.has(id)) {
        delete commentsById.value[id];
      }
    }
  }

  function loadPage(
    postId: number,
    page: number,
    pageSize: number,
    options: { force?: boolean } = {},
  ) {
    const key = pageKey(postId, page, pageSize);
    const state = ensurePageState(postId, page, pageSize);
    state.accessedAt = Date.now();

    if (!postId) {
      state.loading = false;
      return Promise.resolve();
    }

    if (
      !options.force &&
      state.fetchedAt > 0 &&
      Date.now() - state.fetchedAt < CACHE_MAX_AGE
    ) {
      return Promise.resolve();
    }

    const pendingRequest = pageRequests.get(key);
    if (pendingRequest && !options.force) return pendingRequest;

    if (options.force) pageControllers.get(key)?.abort();
    const controller = new AbortController();
    pageControllers.set(key, controller);
    const hasCachedPage = state.fetchedAt > 0;
    const requestMutationRevision = mutationRevision;
    const requestPageRevision = state.revision;
    state.loading = !hasCachedPage;
    state.error = '';

    const request = (async () => {
      try {
        const result = await getPostComments(
          postId,
          { page, pageSize },
          controller.signal,
        );
        if (controller.signal.aborted) return;
        for (const comment of result.items) {
          // Do not let a request started earlier roll back a local mutation.
          if (
            (commentRevisions.get(comment.id) ?? 0) <= requestMutationRevision
          ) {
            setComment(comment);
          }
        }
        if (state.revision !== requestPageRevision) return;
        state.ids = result.items.map((comment) => comment.id);
        state.total = result.total;
        state.fetchedAt = Date.now();
        state.accessedAt = state.fetchedAt;
        prunePageCache(key);
      } catch (reason) {
        if (reason instanceof DOMException && reason.name === 'AbortError')
          return;
        if (!hasCachedPage) {
          state.error =
            reason instanceof Error ? reason.message : '无法加载评论';
        }
      } finally {
        if (pageControllers.get(key) === controller) {
          state.loading = false;
          pageControllers.delete(key);
          pageRequests.delete(key);
        }
      }
    })();
    pageRequests.set(key, request);
    return request;
  }

  async function createComment(
    postId: number,
    input: { content: string; rootId?: number },
  ) {
    const comment = await createPostComment(postId, input);
    setMutatedComment(comment);
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
    pageRequests.delete(key);

    const currentState = ensurePageState(comment.postId, currentPage, pageSize);
    currentState.loading = false;
    currentState.error = '';

    const states = Object.values(pages.value).filter(
      (state) => state.postId === comment.postId,
    );
    for (const state of states) {
      // Keep optimistic content visible, but revalidate server-side pagination.
      state.revision += 1;
      state.total += 1;
      if (state.fetchedAt > 0) state.fetchedAt = 1;
    }

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
    setMutatedComment(comment);
    return comment;
  }

  function applyStatus(id: number, status: number) {
    const comment = commentsById.value[id];
    if (!comment) return;
    setMutatedComment({ ...comment, status, content: '' });
  }

  function registerDeletedCommentsByAuthor(authorId: number) {
    for (const comment of Object.values(commentsById.value)) {
      if (comment.authorId === authorId) applyStatus(comment.id, 2);
    }

    // Pending pages may contain this author's comments even if none are cached.
    for (const [key, state] of Object.entries(pages.value)) {
      const affected = state.ids.some(
        (id) => commentsById.value[id]?.authorId === authorId,
      );
      if (!affected && !pageControllers.has(key)) continue;

      pageControllers.get(key)?.abort();
      pageControllers.delete(key);
      pageRequests.delete(key);
      state.revision += 1;
      if (state.fetchedAt > 0) state.fetchedAt = 1;
      state.loading = false;
    }
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
    registerDeletedCommentsByAuthor,
    updateComment,
  };
});
