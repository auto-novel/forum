<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import { authUser, createPostComment, type PostComment } from '@/api';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';

const props = defineProps<{
  postId: number;
  locked: boolean;
}>();

const emit = defineEmits<{
  created: [comment: PostComment];
}>();

const content = ref('');
const submitting = ref(false);
const submitError = ref('');
const draftKey = computed(() =>
  authUser.value
    ? `forum:comment-draft:${props.postId}:${authUser.value.id}`
    : '',
);

function readDraft(key: string) {
  if (!key) return '';
  try {
    return localStorage.getItem(key) ?? '';
  } catch {
    return '';
  }
}

function saveDraft(key: string, value: string) {
  if (!key) return;
  try {
    if (value.trim()) localStorage.setItem(key, value);
    else localStorage.removeItem(key);
  } catch {
    // Draft persistence is optional when browser storage is unavailable.
  }
}

watch(
  draftKey,
  (key) => {
    content.value = readDraft(key);
  },
  { immediate: true },
);

watch(content, (value) => saveDraft(draftKey.value, value));

async function responseErrorMessage(error: unknown) {
  if (error && typeof error === 'object' && 'response' in error) {
    const response = (error as { response?: Response }).response;
    if (response) {
      try {
        const message = await response.text();
        if (message) return message;
      } catch {
        // Fall back to the client error message when the body is unavailable.
      }
    }
  }
  return error instanceof Error ? error.message : '评论发布失败';
}

async function submitComment() {
  const value = content.value.trim();
  if (!value || submitting.value || props.locked) return;
  submitting.value = true;
  submitError.value = '';
  try {
    const comment = await createPostComment(props.postId, { content: value });
    content.value = '';
    saveDraft(draftKey.value, '');
    emit('created', comment);
  } catch (error) {
    submitError.value = await responseErrorMessage(error);
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <section class="mt-5 rounded-sm bg-surface px-4 py-5 sm:px-6">
    <h2 class="font-semibold text-ink">发表评论</h2>

    <div
      v-if="locked"
      class="mt-4 rounded-md border border-orange-200 bg-orange-50 px-4 py-3 text-sm text-orange-700"
    >
      评论区已锁定，暂时无法发表新评论。
    </div>

    <div
      v-else-if="!authUser"
      class="mt-4 rounded-md border border-divider bg-paper px-4 py-4 text-sm text-muted"
    >
      登录后即可参与评论，请使用页面右上角的登录入口。
    </div>

    <form v-else class="mt-4" @submit.prevent="submitComment">
      <MarkdownEditor
        v-model="content"
        mode="comment"
        placeholder="友善交流，分享你的想法…"
        :disabled="submitting"
      />
      <div class="mt-3 flex flex-wrap items-center justify-between gap-3">
        <p class="text-xs text-muted">支持 Markdown，草稿会自动保存在本机。</p>
        <button
          type="submit"
          class="rounded-sm bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="!content.trim() || submitting"
        >
          {{ submitting ? '发布中…' : '发表评论' }}
        </button>
      </div>
      <p v-if="submitError" class="mt-3 text-sm text-red-600" role="alert">
        {{ submitError }}
      </p>
    </form>
  </section>
</template>
