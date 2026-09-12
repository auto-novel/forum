import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

import { getCategories, type CategoryListItem } from '@/api';

const CACHE_MAX_AGE = 15 * 60 * 1000;
const CATEGORY_TITLES: Record<string, string> = {
  novel: '小说讨论',
  guide: '使用指南',
  feedback: '意见反馈',
};
const FALLBACK_CATEGORIES: CategoryListItem[] = [
  { id: 1, slug: 'novel', tags: [] },
  { id: 2, slug: 'guide', tags: [] },
  { id: 3, slug: 'feedback', tags: [] },
];

export const useCategoryStore = defineStore(
  'category',
  () => {
    const items = ref<CategoryListItem[]>([]);
    const fetchedAt = ref(0);
    const loading = ref(false);
    const error = ref('');
    let refreshPromise: Promise<void> | undefined;

    const categoryItems = computed(() =>
      items.value.length ? items.value : FALLBACK_CATEGORIES,
    );
    const categories = computed(() =>
      categoryItems.value.map((category) => ({
        ...category,
        title: CATEGORY_TITLES[category.slug] ?? category.slug,
      })),
    );
    const defaultCategory = computed(() => categories.value[0]);

    function tagsByCategoryId(categoryId: number) {
      return (
        categoryItems.value.find((category) => category.id === categoryId)
          ?.tags ?? []
      );
    }

    async function refresh() {
      if (refreshPromise) return refreshPromise;
      refreshPromise = (async () => {
        loading.value = true;
        error.value = '';
        try {
          const categories = await getCategories();
          if (categories.length) {
            items.value = categories;
            fetchedAt.value = Date.now();
          }
        } catch (reason) {
          error.value =
            reason instanceof Error ? reason.message : '无法加载分类';
        } finally {
          loading.value = false;
          refreshPromise = undefined;
        }
      })();
      return refreshPromise;
    }

    async function initialize() {
      if (!items.value.length) {
        await refresh();
      } else if (Date.now() - fetchedAt.value >= CACHE_MAX_AGE) {
        void refresh();
      }
    }

    return {
      categories,
      defaultCategory,
      loading,
      error,
      fetchedAt,
      initialize,
      refresh,
      tagsByCategoryId,
      items,
    };
  },
  {
    persist: {
      key: 'forum:categories:v1',
      pick: ['items', 'fetchedAt'],
    },
  },
);
