<script setup lang="ts">
import {
  ArticleOutlined,
  LocalOfferOutlined,
  CommentOutlined,
  DashboardOutlined,
} from '@vicons/material';
import {
  AdminSiteSwitcher,
  WebKitApp,
  WebKitLayout,
  type WebKitMenuOption,
  useWebKit,
} from '@novelia/web-kit';
import { darkTheme, NConfigProvider } from 'naive-ui';
import { RouterView, useRoute } from 'vue-router';

const route = useRoute();

const { whoami, theme } = useWebKit();
const { isDark } = theme;

const menuOptions: WebKitMenuOption[] = [
  {
    label: '概览',
    key: 'overview',
    to: { name: 'overview' },
    icon: DashboardOutlined,
  },
  {
    label: '标签管理',
    key: 'categories',
    to: { name: 'categories' },
    icon: LocalOfferOutlined,
  },
  {
    label: '帖子管理',
    key: 'posts',
    to: { name: 'posts' },
    icon: ArticleOutlined,
  },
  {
    label: '评论管理',
    key: 'comments',
    to: { name: 'comments' },
    icon: CommentOutlined,
  },
  {
    type: 'divider',
    key: 'theme-divider',
  },
];
</script>

<template>
  <NConfigProvider :theme="isDark ? darkTheme : null">
    <WebKitApp>
      <WebKitLayout
        :navigation-options="menuOptions"
        :account-options="[]"
        :selected-navigation-key="String(route.name ?? '')"
      >
        <template #header-actions>
          <AdminSiteSwitcher current="forum" />
        </template>
        <div class="page-container py-4 md:py-6">
          <RouterView v-if="whoami.isAdmin" />
          <section v-else-if="whoami.isSignedIn" class="py-12 text-center">
            <p class="text-4xl font-semibold text-muted">403</p>
            <h1 class="mt-4 text-xl font-medium text-ink">需要管理员权限</h1>
            <p class="mt-2 text-muted">当前账号没有访问管理后台的权限。</p>
          </section>
          <p v-else class="py-12 text-center text-muted">
            请先通过右上角登录账号，再访问管理后台。
          </p>
        </div>
      </WebKitLayout>
    </WebKitApp>
  </NConfigProvider>
</template>
