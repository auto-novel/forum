import type { WebKitContext } from '@novelia/web-kit';
import { inject, type InjectionKey } from 'vue';

export interface Page<T> {
  total: number;
  items: T[];
}

export type PostStatus = 0 | 1 | 2;
export type CommentStatus = 0 | 1 | 2;
export type CommentStatusName = 'published' | 'hidden' | 'deleted';

export interface PostTag {
  id: number;
  name: string;
  color: number;
}

export interface CategoryTag extends PostTag {
  sortOrder: number;
}

export interface CategoryListItem {
  id: number;
  slug: string;
  tags: CategoryTag[];
}

export interface PostSummary {
  id: number;
  categoryId: number;
  title: string;
  authorId: number;
  authorUsername: string;
  authorBlocked: boolean;
  status: PostStatus;
  viewsCount: number;
  commentsCount: number;
  commentsLocked: boolean;
  pinOrder: number | null;
  favorited: boolean;
  createdAt: string;
  updatedAt: string;
  activeAt: string;
  tags: PostTag[];
}

export interface Post extends PostSummary {
  content: string;
}

export type PostSort = 'active' | 'updated' | 'views' | 'comments';

export interface PostComment {
  id: number;
  postId: number;
  rootId: number | null;
  content: string;
  authorId: number;
  authorUsername: string;
  authorBlocked: boolean;
  status: CommentStatus;
  createdAt: string;
  updatedAt: string;
  replyCount: number;
  replies?: Page<PostComment>;
}

export interface BlacklistEntry {
  userId: number;
  createdAt: string;
}

export function createForumApi(
  client: ReturnType<WebKitContext['createClient']>,
) {
  return {
    getPosts(
      params: {
        page: number;
        pageSize: number;
        category?: string;
        query?: string;
        tagIds?: number[];
        sort?: PostSort;
      },
      signal?: AbortSignal,
    ) {
      return client
        .get('post/', {
          searchParams: {
            page: params.page,
            page_size: params.pageSize,
            category: params.category,
            q: params.query,
            tag: params.tagIds?.join(','),
            sort: params.sort,
          },
          signal,
        })
        .json<Page<PostSummary>>();
    },

    getFavoritePosts(
      params: { page: number; pageSize: number },
      signal?: AbortSignal,
    ) {
      return client
        .get('me/favorite', {
          searchParams: {
            page: params.page,
            page_size: params.pageSize,
          },
          signal,
        })
        .json<Page<PostSummary>>();
    },

    getMyPosts(
      params: { page: number; pageSize: number },
      signal?: AbortSignal,
    ) {
      return client
        .get('me/post', {
          searchParams: {
            page: params.page,
            page_size: params.pageSize,
          },
          signal,
        })
        .json<Page<PostSummary>>();
    },

    getPost(id: number, signal?: AbortSignal) {
      return client.get(`post/${id}/`, { signal }).json<Post>();
    },

    async setPostFavorite(id: number, favorited: boolean) {
      const path = `post/${id}/favorite`;
      await (favorited ? client.put(path) : client.delete(path));
    },

    getCategories(signal?: AbortSignal) {
      return client.get('category/', { signal }).json<CategoryListItem[]>();
    },

    createPost(input: {
      categoryId: number;
      title: string;
      content: string;
      tagIds: number[];
    }) {
      return client.post('post/', { json: input }).json<Post>();
    },

    updatePost(
      id: number,
      input: {
        categoryId: number;
        title: string;
        content: string;
        tagIds: number[];
      },
    ) {
      return client.patch(`post/${id}/`, { json: input }).json<Post>();
    },

    async deletePost(id: number) {
      await client.delete(`post/${id}/`);
    },

    async setPostStatus(id: number, status: PostStatus) {
      await client.put(`admin/post/${id}/status`, { json: { status } });
    },

    async lockPost(id: number) {
      await client.put(`admin/post/${id}/lock`);
    },

    async unlockPost(id: number) {
      await client.delete(`admin/post/${id}/lock`);
    },

    async pinPost(id: number, pinOrder: number) {
      await client.put(`admin/post/${id}/pin`, { json: { pinOrder } });
    },

    async unpinPost(id: number) {
      await client.delete(`admin/post/${id}/pin`);
    },

    getPostComments(
      id: number,
      params: { page: number; pageSize: number },
      signal?: AbortSignal,
    ) {
      return client
        .get(`post/${id}/comment`, {
          searchParams: {
            page: params.page,
            page_size: params.pageSize,
          },
          signal,
        })
        .json<Page<PostComment>>();
    },

    getPostCommentReplies(
      postId: number,
      rootId: number,
      params: { page: number; pageSize: number },
      signal?: AbortSignal,
    ) {
      return client
        .get(`post/${postId}/comment/${rootId}/reply`, {
          searchParams: {
            page: params.page,
            page_size: params.pageSize,
          },
          signal,
        })
        .json<Page<PostComment>>();
    },

    createPostComment(id: number, input: { content: string; rootId?: number }) {
      return client
        .post(`post/${id}/comment`, { json: input })
        .json<PostComment>();
    },

    updatePostComment(id: number, content: string) {
      return client
        .patch(`comment/${id}`, { json: { content } })
        .json<PostComment>();
    },

    async deletePostComment(id: number) {
      await client.delete(`comment/${id}`);
    },

    async setPostCommentStatus(id: number, status: CommentStatusName) {
      await client.put(`admin/comment/${id}/status`, { json: { status } });
    },

    async deleteCommentsByAuthor(authorId: number) {
      await client.delete(`admin/comment/author/${authorId}`);
    },

    getBlacklist(signal?: AbortSignal) {
      return client
        .get('me/blacklist', { signal })
        .json<Page<BlacklistEntry>>();
    },

    async setUserBlocked(userId: number, blocked: boolean) {
      const path = `me/blacklist/${userId}`;
      await (blocked ? client.put(path) : client.delete(path));
    },
  };
}

export type ForumApi = ReturnType<typeof createForumApi>;

export const forumApiKey: InjectionKey<ForumApi> = Symbol('ForumApi');

export function useForumApi(): ForumApi {
  const api = inject(forumApiKey);
  if (!api) throw new Error('ForumApi has not been provided');
  return api;
}
