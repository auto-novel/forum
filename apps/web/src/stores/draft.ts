import { computed } from 'vue';
import { useLocalStorage } from '@novelia/web-kit';
import { defineStore } from 'pinia';

const MAX_DRAFT_COUNT = 20;
const DRAFTS_KEY = 'forum:drafts:v1';

export interface PostDraft {
  title: string;
  category: string;
  tagIds: number[];
  content: string;
}

interface DraftState {
  postDrafts: Record<number, PostDraft>;
  commentDrafts: Record<string, string>;
  draftUpdatedAt: Record<string, number>;
}

function emptyDrafts(): DraftState {
  return { postDrafts: {}, commentDrafts: {}, draftUpdatedAt: {} };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function validateDrafts(state: unknown): DraftState | undefined {
  if (
    !isRecord(state) ||
    !isRecord(state.postDrafts) ||
    !isRecord(state.commentDrafts) ||
    !isRecord(state.draftUpdatedAt) ||
    !Object.values(state.postDrafts).every(
      (draft) =>
        isRecord(draft) &&
        typeof draft.title === 'string' &&
        typeof draft.category === 'string' &&
        typeof draft.content === 'string' &&
        Array.isArray(draft.tagIds) &&
        draft.tagIds.every((id) => Number.isSafeInteger(id) && id > 0),
    ) ||
    !Object.values(state.commentDrafts).every(
      (content) => typeof content === 'string',
    ) ||
    !Object.values(state.draftUpdatedAt).every(
      (timestamp) =>
        typeof timestamp === 'number' && Number.isFinite(timestamp),
    )
  ) {
    return undefined;
  }
  return {
    postDrafts: state.postDrafts as Record<number, PostDraft>,
    commentDrafts: state.commentDrafts as Record<string, string>,
    draftUpdatedAt: state.draftUpdatedAt as Record<string, number>,
  };
}

export const useDraftStore = defineStore('draft', () => {
  const drafts = useLocalStorage<DraftState>(DRAFTS_KEY, {
    fallback: emptyDrafts,
    validate: validateDrafts,
  });
  const postDrafts = computed(() => drafts.value.postDrafts);
  const commentDrafts = computed(() => drafts.value.commentDrafts);
  const draftUpdatedAt = computed(() => drafts.value.draftUpdatedAt);

  function postTimestampKey(userId: number | string) {
    return `post:${userId}`;
  }

  function commentTimestampKey(key: string) {
    return `comment:${key}`;
  }

  function trimDrafts() {
    const entries = [
      ...Object.keys(postDrafts.value).map((key) => ({
        type: 'post' as const,
        key,
        updatedAt: draftUpdatedAt.value[postTimestampKey(key)] ?? 0,
      })),
      ...Object.keys(commentDrafts.value).map((key) => ({
        type: 'comment' as const,
        key,
        updatedAt: draftUpdatedAt.value[commentTimestampKey(key)] ?? 0,
      })),
    ].sort((left, right) => left.updatedAt - right.updatedAt);

    for (const entry of entries.slice(0, -MAX_DRAFT_COUNT)) {
      if (entry.type === 'post') delete postDrafts.value[Number(entry.key)];
      else delete commentDrafts.value[entry.key];
      delete draftUpdatedAt.value[
        entry.type === 'post'
          ? postTimestampKey(entry.key)
          : commentTimestampKey(entry.key)
      ];
    }
  }

  function getPostDraft(userId: number) {
    trimDrafts();
    const draft = postDrafts.value[userId];
    if (!draft) return;
    return { ...draft, tagIds: [...draft.tagIds] };
  }

  function savePostDraft(userId: number, draft: PostDraft) {
    if (!draft.title.trim() && !draft.content.trim()) {
      clearPostDraft(userId);
      return;
    }
    postDrafts.value[userId] = { ...draft, tagIds: [...draft.tagIds] };
    draftUpdatedAt.value[postTimestampKey(userId)] = Date.now();
    trimDrafts();
  }

  function clearPostDraft(userId: number) {
    delete postDrafts.value[userId];
    delete draftUpdatedAt.value[postTimestampKey(userId)];
  }

  function getCommentDraft(key: string) {
    trimDrafts();
    return commentDrafts.value[key] ?? '';
  }

  function saveCommentDraft(key: string, content: string) {
    if (!content.trim()) {
      clearCommentDraft(key);
      return;
    }
    commentDrafts.value[key] = content;
    draftUpdatedAt.value[commentTimestampKey(key)] = Date.now();
    trimDrafts();
  }

  function clearCommentDraft(key: string) {
    delete commentDrafts.value[key];
    delete draftUpdatedAt.value[commentTimestampKey(key)];
  }

  function clearAllDrafts() {
    drafts.value = emptyDrafts();
  }

  return {
    clearAllDrafts,
    clearCommentDraft,
    clearPostDraft,
    commentDrafts,
    draftUpdatedAt,
    drafts,
    getCommentDraft,
    getPostDraft,
    postDrafts,
    saveCommentDraft,
    savePostDraft,
  };
});
