<script setup lang="ts">
import {
  CommentOutlined,
  LockOutlined,
  MoreHorizOutlined,
  PushPinOutlined,
  SearchOutlined,
  VisibilityOutlined,
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
  NPagination,
  NSelect,
  NSkeleton,
  NSwitch,
  NTag,
  NText,
  type SelectOption,
} from 'naive-ui';
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useForumApi, type Category, type Post } from '@/api';

const PAGE_SIZE = 20;
const api = useForumApi();
const router = useRouter();
const categories = ref<Category[]>([]);
const posts = ref<Post[]>([]);
const loading = ref(true);
const saving = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const page = ref(1);
const total = ref(0);
const queryInput = ref('');
const categoryInput = ref<string | null>(null);
const query = ref('');
const category = ref('');
const moderationModalOpen = ref(false);
const selectedPost = ref<Post>();
const moderationForm = reactive({
  status: 0,
  commentsLocked: false,
  pinOrder: null as number | null,
});

const categoryOptions = computed<SelectOption[]>(() =>
  categories.value.map((item) => ({ label: item.slug, value: item.slug })),
);
const categoryMap = computed(
  () => new Map(categories.value.map((item) => [item.id, item.slug])),
);
const pageCount = computed(() =>
  Math.max(1, Math.ceil(total.value / PAGE_SIZE)),
);

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}

async function loadPosts() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await api.getPosts({
      page: page.value,
      pageSize: PAGE_SIZE,
      query: query.value,
      category: category.value,
    });
    posts.value = result.items;
    total.value = result.total;
  } catch (error) {
    posts.value = [];
    total.value = 0;
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    loading.value = false;
  }
}

function search() {
  query.value = queryInput.value.trim();
  category.value = categoryInput.value ?? '';
  page.value = 1;
  void loadPosts();
}

function resetFilters() {
  queryInput.value = '';
  categoryInput.value = null;
  query.value = '';
  category.value = '';
  page.value = 1;
  void loadPosts();
}

function openModeration(post: Post) {
  selectedPost.value = post;
  Object.assign(moderationForm, {
    status: post.status,
    commentsLocked: post.commentsLocked,
    pinOrder: post.pinOrder ?? null,
  });
  moderationModalOpen.value = true;
}

async function saveModeration() {
  if (!selectedPost.value) return;
  saving.value = true;
  errorMessage.value = '';
  try {
    await api.moderatePost(selectedPost.value.id, moderationForm);
    moderationModalOpen.value = false;
    successMessage.value = `帖子「${selectedPost.value.title}」已更新`;
    await loadPosts();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    saving.value = false;
  }
}

async function initialize() {
  try {
    categories.value = await api.getCategories();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  }
  await loadPosts();
}

onMounted(initialize);
</script>

<template>
  <main class="posts-page">
    <section class="page-intro">
      <div>
        <n-text tag="h1" class="page-title">帖子管理</n-text>
        <n-text depth="3">检索社区内容，处理置顶、锁评和可见状态。</n-text>
      </div>
      <n-text depth="3" class="total-count">{{ total }} 篇已发布帖子</n-text>
    </section>

    <n-card size="small" class="filter-card">
      <div class="filters">
        <n-input
          v-model:value="queryInput"
          clearable
          placeholder="搜索标题或正文"
          @keyup.enter="search"
        >
          <template #prefix><n-icon :component="SearchOutlined" /></template>
        </n-input>
        <n-select
          v-model:value="categoryInput"
          clearable
          placeholder="全部分类"
          :options="categoryOptions"
        />
        <n-button type="primary" @click="search">筛选</n-button>
        <n-button quaternary @click="resetFilters">重置</n-button>
      </div>
    </n-card>

    <n-alert
      v-if="successMessage"
      type="success"
      closable
      @close="successMessage = ''"
    >
      {{ successMessage }}
    </n-alert>
    <n-alert v-if="errorMessage" type="error" title="帖子列表加载失败">
      <div class="alert-content">
        <n-text>{{ errorMessage }}</n-text>
        <n-button size="small" @click="loadPosts">重新加载</n-button>
      </div>
    </n-alert>

    <div v-if="loading && !posts.length" class="post-list">
      <n-card v-for="index in 5" :key="index" class="post-card">
        <n-skeleton text :repeat="3" />
      </n-card>
    </div>
    <n-empty
      v-else-if="!posts.length"
      class="empty-state"
      :description="query || category ? '没有符合条件的帖子' : '暂无帖子'"
    >
      <template v-if="query || category" #extra>
        <n-button size="small" @click="resetFilters">清除筛选</n-button>
      </template>
    </n-empty>
    <div v-else class="post-list">
      <n-card v-for="post in posts" :key="post.id" class="post-card">
        <div class="post-layout">
          <div class="post-content">
            <div class="post-badges">
              <n-tag size="small" :bordered="false">
                {{
                  categoryMap.get(post.categoryId) ?? `分类 ${post.categoryId}`
                }}
              </n-tag>
              <n-tag
                v-if="post.pinOrder != null"
                size="small"
                type="warning"
                :bordered="false"
              >
                <template #icon>
                  <n-icon :component="PushPinOutlined" />
                </template>
                置顶 {{ post.pinOrder }}
              </n-tag>
              <n-tag
                v-if="post.commentsLocked"
                size="small"
                type="error"
                :bordered="false"
              >
                <template #icon><n-icon :component="LockOutlined" /></template>
                已锁评
              </n-tag>
            </div>
            <n-text strong class="post-title">{{ post.title }}</n-text>
            <n-text depth="3" class="post-excerpt">
              {{ post.content.replace(/\s+/g, ' ').slice(0, 150) }}
            </n-text>
            <div class="post-footer">
              <n-text depth="3">
                {{ post.authorUsername }} · {{ formatDate(post.activeAt) }}
              </n-text>
              <div class="post-metrics">
                <span>
                  <n-icon :component="VisibilityOutlined" />
                  {{ post.viewsCount }}
                </span>
                <span>
                  <n-icon :component="CommentOutlined" />
                  {{ post.commentsCount }}
                </span>
                <n-button
                  text
                  type="primary"
                  size="small"
                  @click="
                    router.push({ name: 'comments', query: { post: post.id } })
                  "
                >
                  审核评论
                </n-button>
              </div>
            </div>
          </div>
          <n-button circle quaternary @click="openModeration(post)">
            <template #icon><n-icon :component="MoreHorizOutlined" /></template>
          </n-button>
        </div>
      </n-card>
    </div>

    <div v-if="total > PAGE_SIZE" class="pagination">
      <n-pagination
        v-model:page="page"
        :page-count="pageCount"
        show-quick-jumper
        @update:page="loadPosts"
      />
    </div>

    <n-modal
      v-model:show="moderationModalOpen"
      preset="card"
      title="管理帖子"
      class="form-modal"
    >
      <n-text v-if="selectedPost" strong class="modal-post-title">
        {{ selectedPost.title }}
      </n-text>
      <n-form label-placement="top">
        <n-form-item label="帖子状态">
          <n-select
            v-model:value="moderationForm.status"
            :options="[
              { label: '正常发布', value: 0 },
              { label: '隐藏', value: 1 },
              { label: '删除', value: 2 },
            ]"
          />
        </n-form-item>
        <n-form-item label="置顶顺序">
          <n-input-number
            v-model:value="moderationForm.pinOrder"
            clearable
            placeholder="留空表示不置顶"
          />
        </n-form-item>
        <n-form-item label="锁定评论区">
          <n-switch v-model:value="moderationForm.commentsLocked" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-actions">
          <n-button @click="moderationModalOpen = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="saveModeration">
            保存设置
          </n-button>
        </div>
      </template>
    </n-modal>
  </main>
</template>

<style scoped>
.posts-page {
  max-width: 980px;
  margin-inline: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.page-intro,
.alert-content,
.post-layout,
.post-footer,
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
.total-count {
  flex: none;
  font-variant-numeric: tabular-nums;
}
.filter-card :deep(.n-card__content) {
  padding: 12px;
}
.filters {
  display: grid;
  grid-template-columns: minmax(200px, 1fr) 180px auto auto;
  gap: 10px;
}
.post-list {
  display: grid;
  gap: 10px;
}
.post-card {
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}
.post-card:hover {
  border-color: rgba(24, 160, 88, 0.4);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.04);
}
.post-layout {
  align-items: flex-start;
}
.post-content {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
}
.post-badges,
.post-metrics {
  display: flex;
  align-items: center;
  gap: 7px;
}
.post-title {
  font-size: 17px;
  line-height: 1.4;
}
.post-excerpt {
  overflow: hidden;
  display: -webkit-box;
  line-height: 1.6;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
.post-footer {
  margin-top: 3px;
}
.post-metrics span {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--n-text-color-3);
  font-size: 12px;
}
.pagination {
  display: flex;
  justify-content: center;
  padding-top: 4px;
}
.empty-state {
  padding: 64px 16px;
}
.form-modal {
  width: min(520px, calc(100vw - 32px));
}
.modal-post-title {
  display: block;
  margin-bottom: 18px;
}
.modal-actions {
  justify-content: flex-end;
}
@media (max-width: 680px) {
  .filters {
    grid-template-columns: minmax(0, 1fr) auto;
  }
  .filters :deep(.n-select) {
    grid-column: 1 / -1;
    grid-row: 2;
  }
  .post-footer {
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }
}
</style>
