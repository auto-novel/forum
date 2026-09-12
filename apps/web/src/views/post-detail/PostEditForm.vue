<script setup lang="ts">
import { computed, ref } from 'vue';

import { updatePost, type Post } from '@/api';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';
import { useCategoryStore } from '@/stores/category';

const props = defineProps<{ post: Post }>();
const emit = defineEmits<{ cancel: []; saved: [post: Post] }>();
const categoryStore = useCategoryStore();

const title = ref(props.post.title);
const content = ref(props.post.content);
const tags = computed(() =>
  categoryStore.tagsByCategoryId(props.post.categoryId),
);
const validTagIds = new Set(tags.value.map((tag) => tag.id));
const selectedTagIds = ref(
  props.post.tags.map((tag) => tag.id).filter((id) => validTagIds.has(id)),
);
const submitting = ref(false);
const submitError = ref('');

const canSubmit = computed(
  () =>
    Boolean(title.value.trim() && content.value.trim()) && !submitting.value,
);

async function errorMessage(reason: unknown) {
  if (reason && typeof reason === 'object' && 'response' in reason) {
    const response = (reason as { response?: Response }).response;
    if (response) {
      try {
        return (await response.text()) || '更新帖子失败';
      } catch {
        // Use the fallback below.
      }
    }
  }
  return reason instanceof Error ? reason.message : '更新帖子失败';
}

async function save() {
  if (!canSubmit.value) return;
  submitting.value = true;
  submitError.value = '';
  try {
    const post = await updatePost(props.post.id, {
      title: title.value.trim(),
      content: content.value,
      tagIds: selectedTagIds.value,
    });
    emit('saved', post);
  } catch (reason) {
    submitError.value = await errorMessage(reason);
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <section class="rounded-sm bg-surface px-4 py-5 sm:px-6 sm:py-7">
    <h1 class="text-xl font-bold text-ink">编辑帖子</h1>
    <form class="mt-5 space-y-5" @submit.prevent="save">
      <div>
        <label for="edit-post-title" class="mb-2 block text-sm font-semibold">
          标题
        </label>
        <input
          id="edit-post-title"
          v-model="title"
          maxlength="500"
          class="block min-h-10 w-full rounded-md border border-border bg-surface px-3 text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
          :disabled="submitting"
          required
        />
      </div>

      <fieldset>
        <legend class="mb-2 text-sm font-semibold">标签</legend>
        <div v-if="tags.length" class="flex flex-wrap gap-2">
          <label
            v-for="tag in tags"
            :key="tag.id"
            class="cursor-pointer rounded-sm border px-3 py-1.5 text-xs font-medium transition-colors"
            :class="
              selectedTagIds.includes(tag.id)
                ? 'border-primary bg-primary-soft text-primary'
                : 'border-border text-muted hover:border-primary hover:text-primary'
            "
          >
            <input
              v-model="selectedTagIds"
              type="checkbox"
              class="sr-only"
              :value="tag.id"
              :disabled="submitting"
            />
            {{ tag.name }}
          </label>
        </div>
        <p v-else class="text-sm text-muted">这个分类暂时没有可用标签。</p>
      </fieldset>

      <div>
        <label class="mb-2 block text-sm font-semibold">正文</label>
        <MarkdownEditor
          v-model="content"
          mode="article"
          :rows="14"
          :disabled="submitting"
        />
      </div>

      <div class="flex justify-end gap-3 border-t border-divider pt-5">
        <button
          type="button"
          class="rounded-sm border border-border px-4 py-2 text-sm font-medium hover:border-primary hover:text-primary"
          :disabled="submitting"
          @click="emit('cancel')"
        >
          取消
        </button>
        <button
          type="submit"
          class="rounded-sm bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="!canSubmit"
        >
          {{ submitting ? '保存中…' : '保存修改' }}
        </button>
      </div>
      <p v-if="submitError" class="text-sm text-red-600" role="alert">
        {{ submitError }}
      </p>
    </form>
  </section>
</template>
