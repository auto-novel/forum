<script setup lang="ts">
import { AddOutlined } from '@vicons/material';
import { NAlert, NButton, NIcon, NSpace, NText } from 'naive-ui';
import { computed, onMounted, reactive, ref, watch } from 'vue';

import { useForumApi, type Category, type Tag } from '@/api';

import CategoryFormModal from './CategoryFormModal.vue';
import CategoryList from './CategoryList.vue';
import TagFormModal from './TagFormModal.vue';
import TagList from './TagList.vue';

const api = useForumApi();
const categories = ref<Category[]>([]);
const tags = ref<Tag[]>([]);
const selectedCategoryId = ref<number>();
const loading = ref(true);
const tagsLoading = ref(false);
const saving = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const categoryModalOpen = ref(false);
const tagModalOpen = ref(false);
const editingCategoryId = ref<number>();
const editingTagId = ref<number>();
const categoryForm = reactive({ slug: '', bannerUrl: '' });
const tagForm = reactive({ name: '', color: 0, isActive: true, sortOrder: 0 });

const selectedCategory = computed(() =>
  categories.value.find((category) => category.id === selectedCategoryId.value),
);

async function loadCategories() {
  loading.value = true;
  errorMessage.value = '';
  try {
    categories.value = await api.getCategories();
    if (
      selectedCategoryId.value == null ||
      !categories.value.some((item) => item.id === selectedCategoryId.value)
    ) {
      selectedCategoryId.value = categories.value[0]?.id;
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    loading.value = false;
  }
}

async function loadTags(categoryId?: number) {
  if (categoryId == null) {
    tags.value = [];
    return;
  }
  tagsLoading.value = true;
  try {
    tags.value = await api.getTags(categoryId);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    tagsLoading.value = false;
  }
}

function openCreateCategory() {
  editingCategoryId.value = undefined;
  categoryForm.slug = '';
  categoryForm.bannerUrl = '';
  categoryModalOpen.value = true;
}

function openEditCategory(category: Category) {
  editingCategoryId.value = category.id;
  categoryForm.slug = category.slug;
  categoryForm.bannerUrl = category.bannerUrl ?? '';
  categoryModalOpen.value = true;
}

async function saveCategory() {
  const slug = categoryForm.slug.trim();
  if (!slug) return;
  saving.value = true;
  errorMessage.value = '';
  try {
    const request = {
      slug,
      bannerUrl: categoryForm.bannerUrl.trim() || null,
    };
    const category = editingCategoryId.value
      ? await api.updateCategory(editingCategoryId.value, request)
      : await api.createCategory(request);
    categoryModalOpen.value = false;
    selectedCategoryId.value = category.id;
    successMessage.value = editingCategoryId.value
      ? '分类信息已更新'
      : '分类已创建';
    await loadCategories();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    saving.value = false;
  }
}

function openCreateTag() {
  editingTagId.value = undefined;
  Object.assign(tagForm, { name: '', color: 0, isActive: true, sortOrder: 0 });
  tagModalOpen.value = true;
}

function openEditTag(tag: Tag) {
  editingTagId.value = tag.id;
  Object.assign(tagForm, {
    name: tag.name,
    color: tag.color,
    isActive: tag.isActive,
    sortOrder: tag.sortOrder,
  });
  tagModalOpen.value = true;
}

async function saveTag() {
  if (selectedCategoryId.value == null || !tagForm.name.trim()) return;
  saving.value = true;
  errorMessage.value = '';
  try {
    const request = { ...tagForm, name: tagForm.name.trim() };
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

watch(selectedCategoryId, (id) => void loadTags(id));
onMounted(loadCategories);
</script>

<template>
  <n-space vertical :size="16" class="categories-page">
    <div class="page-actions">
      <n-text depth="3">共 {{ categories.length }} 个分类</n-text>
      <n-button type="primary" @click="openCreateCategory">
        <template #icon><n-icon :component="AddOutlined" /></template>
        新建分类
      </n-button>
    </div>

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

    <div class="workspace-grid">
      <CategoryList
        :categories="categories"
        :loading="loading"
        :selected-id="selectedCategoryId"
        @select="selectedCategoryId = $event"
        @edit="openEditCategory"
      />
      <TagList
        :category="selectedCategory"
        :tags="tags"
        :loading="tagsLoading"
        @create="openCreateTag"
        @edit="openEditTag"
      />
    </div>

    <CategoryFormModal
      v-model:slug="categoryForm.slug"
      v-model:banner-url="categoryForm.bannerUrl"
      :show="categoryModalOpen"
      :editing="editingCategoryId != null"
      :saving="saving"
      @close="categoryModalOpen = false"
      @save="saveCategory"
    />
    <TagFormModal
      v-model:name="tagForm.name"
      v-model:color="tagForm.color"
      v-model:active="tagForm.isActive"
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
  max-width: 1000px;
  margin-inline: auto;
}

.page-actions {
  min-height: 34px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.workspace-grid {
  display: grid;
  grid-template-columns: minmax(250px, 0.75fr) minmax(0, 1.6fr);
  gap: 18px;
}

@media (max-width: 760px) {
  .workspace-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
