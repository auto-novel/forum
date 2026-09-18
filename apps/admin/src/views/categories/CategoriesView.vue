<script setup lang="ts">
import { NAlert, NEmpty, NSkeleton, NSpace } from 'naive-ui';
import { computed, onMounted, reactive, ref } from 'vue';

import { useForumApi, type Category, type Tag } from '@/api';
import { categoryOrder, categoryTitle } from '@/category';

import TagFormModal from './TagFormModal.vue';
import TagList from './TagList.vue';

const api = useForumApi();
const categories = ref<Category[]>([]);
const tagsByCategory = reactive<Record<number, Tag[]>>({});
const tagsLoadingByCategory = reactive<Record<number, boolean>>({});
const selectedCategoryId = ref<number>();
const loading = ref(true);
const saving = ref(false);
const activeUpdatingTagId = ref<number>();
const errorMessage = ref('');
const successMessage = ref('');
const tagModalOpen = ref(false);
const editingTagId = ref<number>();
const tagForm = reactive({ name: '', color: 0, sortOrder: 0 });
const displayedCategories = computed(() =>
  [...categories.value]
    .sort((left, right) => categoryOrder(left.slug) - categoryOrder(right.slug))
    .map((category) => ({
      ...category,
      title: categoryTitle(category.slug),
    })),
);

async function loadCategories() {
  loading.value = true;
  errorMessage.value = '';
  try {
    categories.value = await api.getCategories();
    loading.value = false;
    await Promise.all(
      categories.value.map((category) => loadTags(category.id)),
    );
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    loading.value = false;
  }
}

async function loadTags(categoryId: number) {
  tagsLoadingByCategory[categoryId] = true;
  try {
    tagsByCategory[categoryId] = await api.getTags(categoryId);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    tagsLoadingByCategory[categoryId] = false;
  }
}

function openCreateTag(categoryId: number) {
  selectedCategoryId.value = categoryId;
  editingTagId.value = undefined;
  Object.assign(tagForm, { name: '', color: 0, sortOrder: 0 });
  tagModalOpen.value = true;
}

function openEditTag(categoryId: number, tag: Tag) {
  selectedCategoryId.value = categoryId;
  editingTagId.value = tag.id;
  Object.assign(tagForm, {
    name: tag.name,
    color: tag.color,
    sortOrder: tag.sortOrder,
  });
  tagModalOpen.value = true;
}

async function saveTag() {
  if (selectedCategoryId.value == null || !tagForm.name.trim()) return;
  saving.value = true;
  errorMessage.value = '';
  try {
    const request = {
      name: tagForm.name.trim(),
      color: tagForm.color,
      sortOrder: tagForm.sortOrder,
    };
    if (editingTagId.value) {
      await api.updateTag(
        selectedCategoryId.value,
        editingTagId.value,
        request,
      );
    } else {
      await api.createTag(selectedCategoryId.value, request);
    }
    tagModalOpen.value = false;
    successMessage.value = editingTagId.value ? '标签已更新' : '标签已创建';
    await loadTags(selectedCategoryId.value);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    saving.value = false;
  }
}

async function toggleTagActive(categoryId: number, tag: Tag) {
  activeUpdatingTagId.value = tag.id;
  errorMessage.value = '';
  try {
    if (tag.isActive) {
      await api.deactivateTag(categoryId, tag.id);
    } else {
      await api.activateTag(categoryId, tag.id);
    }
    successMessage.value = tag.isActive ? '标签已停用' : '标签已启用';
    await loadTags(categoryId);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    activeUpdatingTagId.value = undefined;
  }
}

onMounted(loadCategories);
</script>

<template>
  <n-space vertical :size="16" class="categories-page">
    <n-alert
      v-if="successMessage"
      type="success"
      closable
      @close="successMessage = ''"
    >
      {{ successMessage }}
    </n-alert>
    <n-alert
      v-if="errorMessage"
      type="error"
      closable
      @close="errorMessage = ''"
    >
      {{ errorMessage }}
    </n-alert>

    <div v-if="loading" class="skeleton-stack">
      <n-skeleton v-for="index in 3" :key="index" height="120px" />
    </div>
    <n-empty v-else-if="!displayedCategories.length" description="暂无分类" />
    <div v-else class="category-list">
      <TagList
        v-for="category in displayedCategories"
        :key="category.id"
        :category-name="category.title"
        :tags="tagsByCategory[category.id] ?? []"
        :loading="tagsLoadingByCategory[category.id] ?? false"
        :active-updating-id="activeUpdatingTagId"
        @create="openCreateTag(category.id)"
        @edit="openEditTag(category.id, $event)"
        @toggle-active="toggleTagActive(category.id, $event)"
      />
    </div>

    <TagFormModal
      v-model:name="tagForm.name"
      v-model:color="tagForm.color"
      v-model:sort-order="tagForm.sortOrder"
      :show="tagModalOpen"
      :editing="editingTagId != null"
      :saving="saving"
      @close="tagModalOpen = false"
      @save="saveTag"
    />
  </n-space>
</template>

<style scoped>
.categories-page {
  max-width: 840px;
  margin-inline: auto;
}

.category-list,
.skeleton-stack {
  display: grid;
  gap: 16px;
}
</style>
