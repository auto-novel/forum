import { inject, type InjectionKey } from 'vue';

import type { createAdminKit } from '@novelia/admin-kit';

type AuthApi = ReturnType<typeof createAdminKit>['api'];

export interface Page<T> {
  total: number;
  items: T[];
}

export interface Category {
  id: number;
  slug: string;
  bannerUrl?: string;
}

export interface Tag {
  id: number;
  name: string;
  color: number;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface PostTag {
  id: number;
  name: string;
  color: number;
}

export interface Post {
  id: number;
  categoryId: number;
  title: string;
  authorId: number;
  authorUsername: string;
  content: string;
  status: number;
  viewsCount: number;
  commentsCount: number;
  commentsLocked: boolean;
  pinOrder?: number;
  createdAt: string;
  updatedAt: string;
  activeAt: string;
  tags: PostTag[];
}

export interface Comment {
  id: number;
  postId: number;
  rootId?: number;
  content: string;
  authorId: number;
  authorUsername: string;
  status: number;
  createdAt: string;
  updatedAt: string;
}

interface CategoryRequest {
  slug: string;
  bannerUrl: string | null;
}

interface TagRequest {
  name: string;
  color: number;
  isActive: boolean;
  sortOrder: number;
}

interface PostListParams {
  page: number;
  pageSize: number;
  query?: string;
  category?: string;
}

function endpoint(path: string) {
  return new URL(path, new URL('/api/v1/', window.location.origin));
}

export function createForumApi(authApi: AuthApi) {
  const client = authApi.createClient(
    new URL('/api/v1/', window.location.origin).toString(),
  );

  return {
    getCategories() {
      return client.get(endpoint('category/')).json<Category[]>();
    },
    createCategory(request: CategoryRequest) {
      return client
        .post(endpoint('admin/category/'), { json: request })
        .json<Category>();
    },
    updateCategory(id: number, request: CategoryRequest) {
      return client
        .put(endpoint(`admin/category/${id}`), { json: request })
        .json<Category>();
    },
    getTags(categoryId: number) {
      return client.get(endpoint(`category/${categoryId}/tag`)).json<Tag[]>();
    },
    createTag(categoryId: number, request: TagRequest) {
      return client
        .post(endpoint(`admin/category/${categoryId}/tag`), { json: request })
        .json<Tag>();
    },
    updateTag(categoryId: number, id: number, request: TagRequest) {
      return client
        .put(endpoint(`admin/category/${categoryId}/tag/${id}`), {
          json: request,
        })
        .json<Tag>();
    },
    getPosts(params: PostListParams) {
      return client
        .get(endpoint('post/'), {
          searchParams: {
            page: params.page,
            page_size: params.pageSize,
            q: params.query || undefined,
            category: params.category || undefined,
          },
        })
        .json<Page<Post>>();
    },
    setPostStatus(id: number, status: number) {
      return client
        .put(endpoint(`admin/post/${id}/status`), { json: { status } })
        .text();
    },
    lockPost(id: number) {
      return client.put(endpoint(`admin/post/${id}/lock`)).text();
    },
    unlockPost(id: number) {
      return client.delete(endpoint(`admin/post/${id}/lock`)).text();
    },
    pinPost(id: number, pinOrder: number) {
      return client
        .put(endpoint(`admin/post/${id}/pin`), { json: { pinOrder } })
        .text();
    },
    unpinPost(id: number) {
      return client.delete(endpoint(`admin/post/${id}/pin`)).text();
    },
    getComments(postId: number, page: number, pageSize: number) {
      return client
        .get(endpoint(`post/${postId}/comment`), {
          searchParams: { page, page_size: pageSize },
        })
        .json<Page<Comment>>();
    },
    setCommentStatus(id: number, status: 'published' | 'hidden' | 'deleted') {
      return client
        .put(endpoint(`admin/comment/${id}/status`), { json: { status } })
        .text();
    },
  };
}

export type ForumApi = ReturnType<typeof createForumApi>;

export const forumApiKey: InjectionKey<ForumApi> = Symbol('forum-api');

export function useForumApi() {
  const api = inject(forumApiKey);
  if (!api) throw new Error('Forum API 尚未注册');
  return api;
}
