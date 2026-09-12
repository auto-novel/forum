<script setup lang="ts">
import { computed, ref } from 'vue';

import { updatePost, type Post } from '@/api';
import MarkdownEditor from '@/components/markdown/MarkdownEditor.vue';
import TagSelector from '@/components/TagSelector.vue';
import { notifyError, notifySuccess } from '@/notifications';
import { useCategoryStore } from '@/stores/category';
import { getApiErrorMessage } from '@/utils/apiError';

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

const canSubmit = computed(
  () =>
    Boolean(title.value.trim() && content.value.trim()) && !submitting.value,
);

async function save() {
  if (!canSubmit.value) return;
  submitting.value = true;
  try {
    const post = await updatePost(props.post.id, {
      title: title.value.trim(),
      content: content.value,
      tagIds: selectedTagIds.value,
    });
    notifySuccess('帖子修改已保存');
    emit('saved', post);
  } catch (reason) {
    notifyError(await getApiErrorMessage(reason, '更新帖子失败'));
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

      <TagSelector
        v-model="selectedTagIds"
        :tags="tags"
        label="标签"
        :disabled="submitting"
      />

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
    </form>
  </section>
</template>
