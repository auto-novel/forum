import { ref } from 'vue';
import { defineStore } from 'pinia';

export interface PostDraft {
  title: string;
  category: string;
  tagIds: number[];
  content: string;
}

const MAX_DRAFT_COUNT = 20;

function removeLegacyDraft(key: string) {
  try {
    localStorage.removeItem(key);
  } catch {
    // Legacy draft cleanup is optional when browser storage is unavailable.
  }
}

function readLegacyPostDraft(key: string): PostDraft | undefined {
  try {
    const value = JSON.parse(localStorage.getItem(key) ?? 'null') as unknown;
    if (!value || typeof value !== 'object') return;
    const draft = value as Partial<PostDraft>;
    if (
      typeof draft.title !== 'string' ||
      typeof draft.category !== 'string' ||
      typeof draft.content !== 'string' ||
      !Array.isArray(draft.tagIds)
    )
      return;
    return {
      title: draft.title,
      category: draft.category,
      content: draft.content,
      tagIds: draft.tagIds.filter(
        (id): id is number => Number.isSafeInteger(id) && id > 0,
      ),
    };
  } catch {
    return;
  }
}

function readLegacyCommentDraft(key: string) {
  try {
    return localStorage.getItem(key) ?? '';
  } catch {
    return '';
  }
}

export const useDraftStore = defineStore(
  'draft',
  () => {
    const postDrafts = ref<Record<number, PostDraft>>({});
    const commentDrafts = ref<Record<string, string>>({});
    const draftUpdatedAt = ref<Record<string, number>>({});

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
      let draft: PostDraft | undefined = postDrafts.value[userId];
      if (!draft) {
        const legacyKey = `forum:post-draft:${userId}`;
        draft = readLegacyPostDraft(legacyKey);
        if (draft) {
          postDrafts.value[userId] = draft;
          draftUpdatedAt.value[postTimestampKey(userId)] = Date.now();
          trimDrafts();
        }
        removeLegacyDraft(legacyKey);
      }
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
      removeLegacyDraft(`forum:post-draft:${userId}`);
      trimDrafts();
    }

    function clearPostDraft(userId: number) {
      delete postDrafts.value[userId];
      delete draftUpdatedAt.value[postTimestampKey(userId)];
      removeLegacyDraft(`forum:post-draft:${userId}`);
    }

    function getCommentDraft(key: string) {
      trimDrafts();
      if (commentDrafts.value[key] != null) return commentDrafts.value[key];
      const legacyKey = `forum:comment-draft:${key}`;
      const content = readLegacyCommentDraft(legacyKey);
      if (content) {
        commentDrafts.value[key] = content;
        draftUpdatedAt.value[commentTimestampKey(key)] = Date.now();
        trimDrafts();
      }
      removeLegacyDraft(legacyKey);
      return content;
    }

    function saveCommentDraft(key: string, content: string) {
      if (!content.trim()) {
        clearCommentDraft(key);
        return;
      }
      commentDrafts.value[key] = content;
      draftUpdatedAt.value[commentTimestampKey(key)] = Date.now();
      removeLegacyDraft(`forum:comment-draft:${key}`);
      trimDrafts();
    }

    function clearCommentDraft(key: string) {
      delete commentDrafts.value[key];
      delete draftUpdatedAt.value[commentTimestampKey(key)];
      removeLegacyDraft(`forum:comment-draft:${key}`);
    }

    return {
      clearCommentDraft,
      clearPostDraft,
      commentDrafts,
      draftUpdatedAt,
      getCommentDraft,
      getPostDraft,
      postDrafts,
      saveCommentDraft,
      savePostDraft,
    };
  },
  {
    persist: {
      key: 'forum:drafts:v1',
      storage: localStorage,
    },
  },
);
