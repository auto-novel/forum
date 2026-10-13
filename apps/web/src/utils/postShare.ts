/**
 * 帖子分享所需的最小字段。这里刻意不依赖 `@/api` 类型，
 * 让分享逻辑可以在不引入接口层的情况下复用于详情页以外的场景。
 */
export interface ShareablePost {
  id: number;
  title: string;
}

/** 帖子的规范分享链接，与站内 `/p/:id` 路由保持一致。 */
export function postShareUrl(
  post: ShareablePost,
  origin = window.location.origin,
): string {
  return new URL(`/p/${post.id}`, origin).toString();
}

/** Markdown 引用文本，标题中的方括号需要转义，避免破坏链接语法。 */
export function postShareMarkdown(
  post: ShareablePost,
  origin = window.location.origin,
): string {
  const url = postShareUrl(post, origin);
  const title = post.title.trim().replace(/[[\]]/g, '\\$&');
  return title ? `[${title}](${url})` : url;
}
