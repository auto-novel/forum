<script setup lang="ts">
import {
  AddOutlined,
  EditOutlined,
  ImageOutlined,
  LocalOfferOutlined,
} from '@vicons/material';
import {
  NAlert,
  NButton,
  NCard,
  NEmpty,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NInputNumber,
  NModal,
  NSkeleton,
  NSwitch,
  NTag,
  NText,
} from 'naive-ui';
import { computed, onMounted, reactive, ref, watch } from 'vue';

import { useForumApi, type Category, type Tag } from '@/api';

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
  <main class="categories-page">
    <section class="page-intro">
      <div>
        <n-text tag="h1" class="page-title">组织讨论空间</n-text>
        <n-text depth="3">维护分类入口、视觉横幅和可选标签。</n-text>
      </div>
      <n-button type="primary" @click="openCreateCategory">
        <template #icon><n-icon :component="AddOutlined" /></template>
        新建分类
      </n-button>
    </section>

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
      <section class="category-panel">
        <n-text strong class="panel-label">分类</n-text>
        <div v-if="loading" class="skeleton-stack">
          <n-skeleton v-for="index in 3" :key="index" height="92px" />
        </div>
        <n-empty v-else-if="!categories.length" description="暂无分类" />
        <button
          v-for="category in categories"
          v-else
          :key="category.id"
          type="button"
          :class="[
            'category-card',
            { selected: selectedCategoryId === category.id },
          ]"
          @click="selectedCategoryId = category.id"
        >
          <span class="category-art">
            <img v-if="category.bannerUrl" :src="category.bannerUrl" alt="" />
            <n-icon v-else :component="ImageOutlined" />
          </span>
          <span class="category-copy">
            <n-text strong>{{ category.slug }}</n-text>
            <n-text depth="3">ID {{ category.id }}</n-text>
          </span>
          <n-button
            text
            aria-label="编辑分类"
            @click.stop="openEditCategory(category)"
          >
            <template #icon><n-icon :component="EditOutlined" /></template>
          </n-button>
        </button>
      </section>

      <n-card class="tag-panel" :bordered="false">
        <template #header>
          <div class="panel-header">
            <div>
              <n-text strong>{{ selectedCategory?.slug ?? '标签' }}</n-text>
              <n-text depth="3" class="panel-caption">
                {{ tags.length }} 个标签
              </n-text>
            </div>
            <n-button
              size="small"
              secondary
              type="primary"
              :disabled="!selectedCategory"
              @click="openCreateTag"
            >
              <template #icon><n-icon :component="AddOutlined" /></template>
              新建标签
            </n-button>
          </div>
        </template>

        <div v-if="tagsLoading" class="skeleton-stack">
          <n-skeleton v-for="index in 4" :key="index" text :repeat="2" />
        </div>
        <n-empty
          v-else-if="!selectedCategory || !tags.length"
          :description="selectedCategory ? '该分类暂无标签' : '请先创建分类'"
        />
        <div v-else class="tag-list">
          <div v-for="tag in tags" :key="tag.id" class="tag-row">
            <span class="tag-symbol">
              <n-icon :component="LocalOfferOutlined" />
            </span>
            <div class="tag-copy">
              <div class="tag-title">
                <n-text strong>{{ tag.name }}</n-text>
                <n-tag
                  :type="tag.isActive ? 'success' : 'default'"
                  size="small"
                >
                  {{ tag.isActive ? '启用' : '停用' }}
                </n-tag>
              </div>
              <n-text depth="3">
                色号 {{ tag.color }} · 排序 {{ tag.sortOrder }}
              </n-text>
            </div>
            <n-button size="small" quaternary @click="openEditTag(tag)">
              编辑
            </n-button>
          </div>
        </div>
      </n-card>
    </div>

    <n-modal
      v-model:show="categoryModalOpen"
      preset="card"
      :title="editingCategoryId ? '编辑分类' : '新建分类'"
      class="form-modal"
    >
      <n-form label-placement="top">
        <n-form-item label="分类标识" required>
          <n-input
            v-model:value="categoryForm.slug"
            placeholder="例如 general"
          />
        </n-form-item>
        <n-form-item label="横幅地址">
          <n-input
            v-model:value="categoryForm.bannerUrl"
            placeholder="https://example.com/banner.webp"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-actions">
          <n-button @click="categoryModalOpen = false">取消</n-button>
          <n-button
            type="primary"
            :loading="saving"
            :disabled="!categoryForm.slug.trim()"
            @click="saveCategory"
          >
            保存
          </n-button>
        </div>
      </template>
    </n-modal>

    <n-modal
      v-model:show="tagModalOpen"
      preset="card"
      :title="editingTagId ? '编辑标签' : '新建标签'"
      class="form-modal"
    >
      <n-form label-placement="top">
        <n-form-item label="标签名称" required>
          <n-input v-model:value="tagForm.name" placeholder="输入标签名称" />
        </n-form-item>
        <div class="form-grid">
          <n-form-item label="预定义色号">
            <n-input-number v-model:value="tagForm.color" :min="0" />
          </n-form-item>
          <n-form-item label="排序值">
            <n-input-number v-model:value="tagForm.sortOrder" />
          </n-form-item>
        </div>
        <n-form-item v-if="editingTagId" label="允许新帖使用">
          <n-switch v-model:value="tagForm.isActive" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-actions">
          <n-button @click="tagModalOpen = false">取消</n-button>
          <n-button
            type="primary"
            :loading="saving"
            :disabled="!tagForm.name.trim()"
            @click="saveTag"
          >
            保存
          </n-button>
        </div>
      </template>
    </n-modal>
  </main>
</template>

<style scoped>
.categories-page {
  max-width: 1120px;
  margin-inline: auto;
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.page-intro,
.panel-header,
.modal-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.page-title {
  display: block;
  margin: 0 0 5px;
  font-size: 23px;
  font-weight: 680;
}
.workspace-grid {
  display: grid;
  grid-template-columns: minmax(250px, 0.75fr) minmax(0, 1.6fr);
  gap: 18px;
}
.category-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.panel-label {
  padding: 0 4px 4px;
  font-size: 13px;
}
.category-card {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--n-border-color);
  border-radius: 12px;
  color: inherit;
  background: var(--n-color);
  display: flex;
  align-items: center;
  gap: 12px;
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.2s,
    box-shadow 0.2s,
    transform 0.2s;
}
.category-card:hover,
.category-card.selected {
  border-color: rgba(24, 160, 88, 0.55);
  box-shadow: 0 7px 22px rgba(24, 160, 88, 0.08);
}
.category-card.selected {
  transform: translateX(3px);
}
.category-art {
  width: 54px;
  height: 54px;
  overflow: hidden;
  border-radius: 10px;
  background: linear-gradient(
    135deg,
    rgba(24, 160, 88, 0.16),
    rgba(32, 128, 240, 0.12)
  );
  display: grid;
  place-items: center;
  flex: none;
  color: #18a058;
  font-size: 22px;
}
.category-art img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.category-copy,
.panel-header > div,
.tag-copy {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 3px;
}
.panel-caption {
  font-size: 12px;
}
.tag-panel {
  min-height: 360px;
}
.tag-list {
  display: grid;
  gap: 2px;
}
.tag-row {
  min-height: 64px;
  padding: 10px 4px;
  border-bottom: 1px solid var(--n-border-color);
  display: flex;
  align-items: center;
  gap: 12px;
}
.tag-row:last-child {
  border-bottom: 0;
}
.tag-symbol {
  width: 34px;
  height: 34px;
  border-radius: 9px;
  color: #2080f0;
  background: rgba(32, 128, 240, 0.1);
  display: grid;
  place-items: center;
  flex: none;
}
.tag-title {
  display: flex;
  align-items: center;
  gap: 8px;
}
.skeleton-stack {
  display: grid;
  gap: 12px;
}
.form-modal {
  width: min(520px, calc(100vw - 32px));
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}
.modal-actions {
  justify-content: flex-end;
}
@media (max-width: 760px) {
  .workspace-grid {
    grid-template-columns: minmax(0, 1fr);
  }
  .page-intro {
    align-items: flex-start;
  }
}
</style>
