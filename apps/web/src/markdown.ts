import { createMarkdown } from '@novelia/web-kit';
import anchor from 'markdown-it-anchor';

export type MarkdownMode = 'article' | 'comment';

const COMMENT_DISABLED_RULES = [
  'blockquote',
  'code',
  'fence',
  'heading',
  'hr',
  'image',
  'lheading',
  'reference',
  'table',
];

const articleMarkdown = createMarkdown().use(anchor, {
  // Keep Unicode IDs so existing percent-encoded Chinese fragments match.
  slugify: (title: string) => title.trim().toLowerCase().replace(/\s+/g, '-'),
});
const commentMarkdown = createMarkdown().disable(COMMENT_DISABLED_RULES);

const markdownByMode = {
  article: articleMarkdown,
  comment: commentMarkdown,
};

/** Returns the shared parser configured for forum articles or comments. */
export function getMarkdown(mode: MarkdownMode) {
  return markdownByMode[mode];
}
