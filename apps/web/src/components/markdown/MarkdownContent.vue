<script setup lang="ts">
import { computed } from 'vue';

import { renderMarkdown, type MarkdownMode } from './renderMarkdown';

const props = defineProps<{
  mode: MarkdownMode;
  source: string;
}>();

const renderedContent = computed(() =>
  renderMarkdown(props.source, props.mode),
);

function spoilerFromTarget(target: EventTarget | null) {
  return target instanceof Element
    ? target.closest<HTMLElement>('[data-markdown-spoiler]')
    : null;
}

function toggleSpoiler(spoilerElement: HTMLElement) {
  spoilerElement.dataset.hide =
    spoilerElement.dataset.hide === 'true' ? 'false' : 'true';
}

function handleClick(event: MouseEvent) {
  const spoilerElement = spoilerFromTarget(event.target);
  if (!spoilerElement) return;
  if (
    spoilerElement.dataset.hide === 'false' &&
    event.target instanceof Element &&
    event.target.closest('a')
  ) {
    return;
  }
  if (spoilerElement.dataset.hide === 'true') event.preventDefault();
  toggleSpoiler(spoilerElement);
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' && event.key !== ' ') return;
  const spoilerElement = spoilerFromTarget(event.target);
  if (!spoilerElement) return;
  event.preventDefault();
  toggleSpoiler(spoilerElement);
}
</script>

<template>
  <div
    class="markdown-content"
    :class="`markdown-content--${mode}`"
    v-html="renderedContent"
    @click="handleClick"
    @keydown="handleKeydown"
  />
</template>

<style scoped>
.markdown-content {
  color: var(--color-ink);
  font-size: 0.9375rem;
  line-height: 1.75rem;
  overflow-wrap: anywhere;
}

.markdown-content--comment {
  font-size: 0.875rem;
  line-height: 1.5rem;
}

.markdown-content :deep(> :first-child) {
  margin-top: 0;
}

.markdown-content :deep(> :last-child) {
  margin-bottom: 0;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4) {
  margin-block: 1.6em 0.65em;
  color: var(--color-ink);
  font-weight: 700;
  line-height: 1.3;
}

.markdown-content :deep(h1) {
  padding-bottom: 0.35em;
  border-bottom: 1px solid var(--color-divider);
  font-size: 1.5em;
}

.markdown-content :deep(h2) {
  padding-bottom: 0.3em;
  border-bottom: 1px solid var(--color-divider);
  font-size: 1.3em;
}

.markdown-content :deep(h3) {
  font-size: 1.15em;
}

.markdown-content :deep(h4) {
  font-size: 1em;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4),
.markdown-content :deep(h5),
.markdown-content :deep(h6) {
  scroll-margin-top: 5rem;
}

.markdown-content :deep(p),
.markdown-content :deep(ul),
.markdown-content :deep(ol),
.markdown-content :deep(blockquote),
.markdown-content :deep(pre),
.markdown-content :deep(table) {
  margin-block: 0 1em;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  padding-left: 1.5rem;
}

.markdown-content :deep(ul) {
  list-style: disc;
}

.markdown-content :deep(ol) {
  list-style: decimal;
}

.markdown-content :deep(li + li) {
  margin-top: 0.25em;
}

.markdown-content :deep(a) {
  color: var(--color-primary);
  text-decoration: underline;
  text-decoration-color: color-mix(
    in srgb,
    var(--color-primary) 45%,
    transparent
  );
  text-underline-offset: 0.2em;
}

.markdown-content :deep(a:hover) {
  color: var(--color-primary-hover);
}

.markdown-content :deep(blockquote) {
  border-left: 3px solid var(--color-primary);
  padding: 0.25rem 0 0.25rem 1rem;
  color: var(--color-muted);
}

.markdown-content :deep(code) {
  border-radius: 0.25rem;
  background: var(--color-paper);
  padding: 0.15em 0.35em;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.88em;
}

.markdown-content :deep(pre) {
  overflow-x: auto;
  border: 1px solid var(--color-divider);
  border-radius: 0.375rem;
  background: var(--color-paper);
  padding: 1rem;
  line-height: 1.6;
}

.markdown-content :deep(pre code) {
  border-radius: 0;
  background: transparent;
  padding: 0;
  font-size: 0.85em;
}

.markdown-content :deep(hr) {
  margin-block: 1.5rem;
  border: 0;
  border-top: 1px solid var(--color-divider);
}

.markdown-content :deep(table) {
  display: block;
  max-width: 100%;
  overflow-x: auto;
  border-collapse: collapse;
}

.markdown-content :deep(th),
.markdown-content :deep(td) {
  border: 1px solid var(--color-border);
  padding: 0.4rem 0.7rem;
  text-align: left;
}

.markdown-content :deep(th) {
  background: var(--color-paper);
  font-weight: 600;
}

.markdown-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 0.375rem;
}

.markdown-content :deep(details) {
  margin-bottom: 1rem;
  border: 1px solid var(--color-border);
  border-radius: 0.375rem;
  padding: 0.65rem 0.85rem;
}

.markdown-content :deep(summary) {
  cursor: pointer;
  font-weight: 600;
}

.markdown-content :deep(details[open] summary) {
  margin-bottom: 0.75rem;
}

.markdown-content :deep(.markdown-star-rating) {
  margin-bottom: 1rem;
  color: #f59e0b;
  font-size: 1rem;
}

.markdown-content :deep([data-markdown-spoiler]) {
  cursor: pointer;
  border-radius: 0.2em;
  background: var(--color-ink);
  padding: 0.05em 0.25em;
  transition: color 150ms ease;
}

.markdown-content :deep([data-markdown-spoiler][data-hide='true']),
.markdown-content :deep([data-markdown-spoiler][data-hide='true'] *) {
  color: transparent !important;
}

.markdown-content :deep([data-markdown-spoiler][data-hide='false']),
.markdown-content :deep([data-markdown-spoiler][data-hide='false'] *),
.markdown-content :deep([data-markdown-spoiler][data-hide='true']:hover),
.markdown-content :deep([data-markdown-spoiler][data-hide='true']:hover *),
.markdown-content :deep([data-markdown-spoiler][data-hide='true']:focus),
.markdown-content :deep([data-markdown-spoiler][data-hide='true']:focus *) {
  color: var(--color-surface);
}
</style>
