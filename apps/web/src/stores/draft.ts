import { useLocalStorage } from '@novelia/web-kit';
import { defineStore } from 'pinia';

const DRAFTS_KEY = 'forum:drafts:v1';

export interface PostDraft {
  title: string;
  category: string;
  tagIds: number[];
  content: string;
}

interface StoredPostDraft {
  userId: number;
  draft: PostDraft;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function parseDraft(value: unknown): PostDraft | undefined {
  if (
    !isRecord(value) ||
    typeof value.title !== 'string' ||
    typeof value.category !== 'string' ||
    typeof value.content !== 'string' ||
    !Array.isArray(value.tagIds) ||
    !value.tagIds.every((id) => Number.isSafeInteger(id) && id > 0)
  )
    return undefined;
  return {
    title: value.title,
    category: value.category,
    content: value.content,
    tagIds: [...value.tagIds],
  };
}

function validateDraft(state: unknown): StoredPostDraft | undefined {
  if (!isRecord(state)) return undefined;
  if ('userId' in state) {
    const draft = parseDraft(state.draft);
    return Number.isSafeInteger(state.userId) &&
      Number(state.userId) > 0 &&
      draft
      ? { userId: Number(state.userId), draft }
      : undefined;
  }

  // 2026-11-11：兼容旧 postDrafts 和按用户存储的映射，取第一份草稿迁移，忽略旧时间戳。
  const entries = 'postDrafts' in state ? state.postDrafts : state;
  if (!isRecord(entries)) return undefined;
  const first = Object.entries(entries)[0];
  if (!first) return undefined;
  const userId = Number(first[0]);
  const draft = parseDraft(first[1]);
  return Number.isSafeInteger(userId) && userId > 0 && draft
    ? { userId, draft }
    : undefined;
}

export const usePostDraftStore = defineStore('post-draft', () => {
  const stored = useLocalStorage<StoredPostDraft | undefined>(DRAFTS_KEY, {
    fallback: () => undefined,
    validate: validateDraft,
  });
  // 初始化时将解析、迁移后的单条草稿写回，后续只处理新格式。
  if (stored.value) stored.value = { ...stored.value };

  function get(userId: number): PostDraft | undefined {
    if (!Number.isSafeInteger(userId) || userId <= 0) return;
    const current = stored.value;
    if (!current || current.userId !== userId) return;
    return { ...current.draft, tagIds: [...current.draft.tagIds] };
  }

  function save(userId: number, draft: PostDraft) {
    if (!Number.isSafeInteger(userId) || userId <= 0) return;
    if (!draft.title.trim() && !draft.content.trim()) {
      clear(userId);
      return;
    }
    stored.value = { userId, draft: { ...draft, tagIds: [...draft.tagIds] } };
  }

  function clear(userId: number) {
    if (get(userId)) clearAll();
  }

  function clearAll() {
    stored.value = undefined;
  }

  return { get, save, clear, clearAll };
});
