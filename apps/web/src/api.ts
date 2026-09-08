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

function endpoint(path: string) {
  return new URL(path, new URL('/api/v1/', window.location.origin));
}

async function getJson<T>(url: URL, signal?: AbortSignal): Promise<T> {
  const response = await fetch(url, {
    headers: { Accept: 'application/json' },
    signal,
  });

  if (!response.ok) {
    throw new Error(`请求失败（${response.status}）`);
  }

  return response.json() as Promise<T>;
}

export function getPosts(
  params: { page: number; pageSize: number; category?: string },
  signal?: AbortSignal,
) {
  const url = endpoint('post/');
  url.searchParams.set('page', String(params.page));
  url.searchParams.set('page_size', String(params.pageSize));
  if (params.category) url.searchParams.set('category', params.category);
  return getJson<Page<Post>>(url, signal);
}
