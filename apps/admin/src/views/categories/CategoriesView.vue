<script setup lang="ts">
import AdminAlert from '@/components/AdminAlert.vue';
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
const tagForm = reactive<{
  name: string;
  color: number | null;
  sortOrder: number | null;
}>({ name: '', color: 0, sortOrder: 0 });
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
  const { name, color, sortOrder } = tagForm;
  if (
    selectedCategoryId.value == null ||
    !name.trim() ||
    color == null ||
    sortOrder == null
  )
    return;
  saving.value = true;
  errorMessage.value = '';
  try {
    const request = {
      name: name.trim(),
      color,
      sortOrder,
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
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-4">
    <header class="pb-2">
      <h1 class="text-xl font-semibold text-ink">标签管理</h1>
      <p class="mt-2 text-sm text-muted">
        按分类维护标签，管理启用状态、色号与排序。
      </p>
    </header>

    <AdminAlert
      v-if="successMessage"
      type="success"
      closable
      @close="successMessage = ''"
    >
      {{ successMessage }}
    </AdminAlert>
    <AdminAlert
      v-if="errorMessage"
      type="error"
      closable
      @close="errorMessage = ''"
    >
      {{ errorMessage }}
    </AdminAlert>

    <div v-if="loading" role="status">
      <span class="sr-only">正在加载分类</span>
      <div aria-hidden="true" class="divide-y divide-divider">
        <div v-for="index in 3" :key="index" class="py-6">
          <div class="mb-4 flex items-center justify-between">
            <div class="skeleton h-5 w-24"></div>
            <div class="skeleton h-8 w-24"></div>
          </div>
          <div class="grid gap-4 border-y border-divider px-3 py-4">
            <div class="skeleton h-4 w-2/3"></div>
            <div class="skeleton h-4 w-1/2"></div>
            <div class="skeleton h-4 w-3/5"></div>
          </div>
        </div>
      </div>
    </div>
    <p
      v-else-if="!displayedCategories.length"
      class="py-8 text-center text-sm text-muted"
    >
      暂无分类
    </p>
    <div v-else class="divide-y divide-divider">
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
  </div>
</template>
