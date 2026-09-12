<script setup lang="ts">
import {
  ArticleOutlined,
  ExploreOutlined,
  ForumOutlined,
  MenuBookOutlined,
  StarBorderOutlined,
} from '@vicons/material';
import { WebKitLayout, type WebKitMenuOption } from '@novelia/web-kit';
import { computed, type Component } from 'vue';
import { RouterView, useRoute } from 'vue-router';

import { useCategoryStore } from '@/stores/category';

const route = useRoute();
const categoryStore = useCategoryStore();

function categoryIcon(slug: string): Component {
  if (slug === 'novel') return MenuBookOutlined;
  if (slug === 'guide') return ExploreOutlined;
  return ForumOutlined;
}

const navigationOptions = computed<WebKitMenuOption[]>(() =>
  categoryStore.categories.map((category) => ({
    key: category.slug,
    label: category.title,
    icon: categoryIcon(category.slug),
    to: { name: 'posts', params: { slug: category.slug } },
  })),
);

const accountOptions: WebKitMenuOption[] = [
  {
    key: 'my-posts',
    label: '我的帖子',
    icon: ArticleOutlined,
    to: { name: 'my-posts' },
  },
  {
    key: 'favorites',
    label: '我的收藏',
    icon: StarBorderOutlined,
    to: { name: 'favorites' },
  },
];

const selectedNavigationKey = computed(() =>
  route.name === 'posts' && typeof route.params.slug === 'string'
    ? route.params.slug
    : undefined,
);
</script>

<template>
  <WebKitLayout
    :navigation-options="navigationOptions"
    :account-options="accountOptions"
    :selected-navigation-key="selectedNavigationKey"
  >
    <RouterView />
  </WebKitLayout>
</template>
