import { createAuthApi, type AuthUser } from '@novelia/auth-api';
import { readonly, ref } from 'vue';

const authUrl = new URL(__AUTH_URL__, window.location.origin);

export const authApi = createAuthApi({
  app: 'f',
  url: authUrl.toString(),
  storage: {
    key: 'f-session',
    target: localStorage,
  },
});

const user = ref<AuthUser>();

authApi.watchUser((profile) => {
  user.value = profile;
});

export const authUser = readonly(user);

const client = authApi.createClient(
  new URL('/api/v1/', window.location.origin).toString(),
);

export interface Page<T> {
  total: number;
  items: T[];
}

export interface Category {
  id: number;
  slug: string;
  title: string;
}

export const CATEGORIES: Category[] = [
  { id: 1, slug: 'novel', title: '小说讨论' },
  { id: 2, slug: 'guide', title: '使用指南' },
  { id: 3, slug: 'feedback', title: '意见反馈' },
];

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

export interface PostComment {
  id: number;
  postId: number;
  rootId: number | null;
  content: string;
  authorId: number;
  authorUsername: string;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export function getPosts(
  params: { page: number; pageSize: number; category?: string },
  signal?: AbortSignal,
) {
  return client
    .get('post/', {
      searchParams: {
        page: params.page,
        page_size: params.pageSize,
        category: params.category,
      },
      signal,
    })
    .json<Page<Post>>();
}

export function getPost(id: number, signal?: AbortSignal) {
  return client.get(`post/${id}/`, { signal }).json<Post>();
}

export function getPostComments(
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
}

export function createPostComment(
  id: number,
  input: { content: string; rootId?: number },
) {
  return client.post(`post/${id}/comment`, { json: input }).json<PostComment>();
}
