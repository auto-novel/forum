import { container } from '@mdit/plugin-container';
import { spoiler } from '@mdit/plugin-spoiler';
import DOMPurify from 'dompurify';
import MarkdownIt, { type RendererRule, type Token } from 'markdown-it';
import MarkdownItAnchor from 'markdown-it-anchor';

export type MarkdownMode = 'article' | 'comment';

const COMMENT_DISABLED_RULES = [
  'backticks',
  'blockquote',
  'code',
  'entity',
  'escape',
  'fence',
  'heading',
  'hr',
  'image',
  'lheading',
  'reference',
  'table',
];

function createMarkdown(mode: MarkdownMode) {
  const markdown = new MarkdownIt({
    html: false,
    linkify: true,
    breaks: true,
  });
  markdown.use(MarkdownItAnchor);
  markdown.use(spoiler, {
    tag: 'span',
    attrs: [
      ['data-markdown-spoiler', ''],
      ['data-hide', 'true'],
      ['role', 'button'],
      ['tabindex', '0'],
    ],
  });
  markdown.use(container, {
    name: 'details',
    validate: (params) => params.trim().split(' ', 2)[0] === 'details',
    openRenderer: (tokens: Token[], index: number): string => {
      const summary = markdown.utils.escapeHtml(
        tokens[index].info.trim().slice(8).trim() || '点击展开',
      );
      return `<details><summary>${summary}</summary>\n`;
    },
    closeRenderer: () => '</details>\n',
  });
  markdown.use(container, {
    name: 'star',
    validate: (params) => params.trim().split(' ', 2)[0] === 'star',
    openRenderer: (tokens: Token[], index: number): string => {
      const rawValue = Number(tokens[index].info.trim().slice(5).trim());
      const value = Number.isFinite(rawValue)
        ? Math.min(5, Math.max(0, rawValue))
        : 0;
      return `<div class="markdown-star-rating" role="img" aria-label="评分 ${value} / 5">★ <strong>${value}</strong> / 5`;
    },
    closeRenderer: () => '</div>\n',
  });

  if (mode === 'comment') markdown.disable(COMMENT_DISABLED_RULES);

  const defaultLinkOpen: RendererRule =
    markdown.renderer.rules.link_open ??
    ((tokens, index, options, _env, renderer) =>
      renderer.renderToken(tokens, index, options));
  const renderExternalLink: RendererRule = (
    tokens,
    index,
    options,
    env,
    renderer,
  ) => {
    const href = tokens[index].attrGet('href');
    if (typeof href === 'string' && !href.startsWith('#')) {
      tokens[index].attrSet('target', '_blank');
      tokens[index].attrSet('rel', 'noopener noreferrer');
    }
    return defaultLinkOpen(tokens, index, options, env, renderer);
  };
  markdown.renderer.rules.link_open = renderExternalLink;

  return markdown;
}

const markdownByMode = {
  article: createMarkdown('article'),
  comment: createMarkdown('comment'),
};

export function renderMarkdown(source: string, mode: MarkdownMode) {
  return DOMPurify.sanitize(markdownByMode[mode].render(source), {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style'],
    FORBID_ATTR: ['style'],
  });
}
