import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

import {
  getFavoritePosts,
  getMyPosts,
  getPost,
  getPosts,
  type Page,
  type Post,
  type PostSort,
  type PostSummary,
} from '@/api';

type ListRequest = (signal: AbortSignal) => Promise<Page<PostSummary>>;

export const usePostStore = defineStore('post', () => {
  const summariesById = ref<Record<number, PostSummary>>({});
  const postsById = ref<Record<number, Post>>({});
  const listPostIds = ref<number[]>([]);
  const listTotal = ref(0);
  const listLoading = ref(false);
  const listError = ref('');
  const currentPostId = ref<number>();
  const detailLoading = ref(false);
  const detailError = ref('');
  let listController: AbortController | undefined;
  let detailController: AbortController | undefined;

  const listPosts = computed(() =>
    listPostIds.value.flatMap((id) => {
      const post = summariesById.value[id];
      return post ? [post] : [];
    }),
  );
  const currentPost = computed(() =>
    currentPostId.value == null
      ? undefined
      : postsById.value[currentPostId.value],
  );

  function setPost(post: Post) {
    postsById.value[post.id] = post;
    summariesById.value[post.id] = post;
  }

  function setPosts(posts: PostSummary[]) {
    for (const post of posts) summariesById.value[post.id] = post;
  }

  function clearList() {
    listController?.abort();
    listController = undefined;
    listPostIds.value = [];
    listTotal.value = 0;
    listLoading.value = false;
    listError.value = '';
  }

  async function loadList(request: ListRequest, fallbackError: string) {
    listController?.abort();
    const controller = new AbortController();
    listController = controller;
    listLoading.value = true;
    listError.value = '';

    try {
      const result = await request(controller.signal);
      setPosts(result.items);
      listPostIds.value = result.items.map((post) => post.id);
      listTotal.value = result.total;
    } catch (reason) {
      if (reason instanceof DOMException && reason.name === 'AbortError')
        return;
      listError.value =
        reason instanceof Error ? reason.message : fallbackError;
      listPostIds.value = [];
      listTotal.value = 0;
    } finally {
      if (listController === controller) listLoading.value = false;
    }
  }

  function loadCategoryPosts(params: {
    page: number;
    pageSize: number;
    category: string;
    query?: string;
    tagIds?: number[];
    sort: PostSort;
  }) {
    return loadList((signal) => getPosts(params, signal), '无法加载帖子');
  }

  function loadMyPosts(params: { page: number; pageSize: number }) {
    return loadList((signal) => getMyPosts(params, signal), '无法加载我的帖子');
  }

  function loadFavoritePosts(params: { page: number; pageSize: number }) {
    return loadList(
      (signal) => getFavoritePosts(params, signal),
      '无法加载收藏帖子',
    );
  }

  async function loadPost(id: number) {
    detailController?.abort();
    detailError.value = '';
    currentPostId.value = id || undefined;

    if (!id) {
      detailLoading.value = false;
      detailError.value = '帖子地址无效';
      return;
    }

    const controller = new AbortController();
    detailController = controller;
    detailLoading.value = true;
    try {
      const post = await getPost(id, controller.signal);
      setPost(post);
      return post;
    } catch (reason) {
      if (reason instanceof DOMException && reason.name === 'AbortError')
        return;
      detailError.value =
        reason instanceof Error ? reason.message : '无法加载帖子';
    } finally {
      if (detailController === controller) detailLoading.value = false;
    }
  }

  return {
    currentPost,
    detailError,
    detailLoading,
    listError,
    listLoading,
    listPosts,
    listTotal,
    clearList,
    loadCategoryPosts,
    loadFavoritePosts,
    loadMyPosts,
    loadPost,
    setPost,
  };
});
