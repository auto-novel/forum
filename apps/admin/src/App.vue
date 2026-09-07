<script setup lang="ts">
import {
  ArticleOutlined,
  CategoryOutlined,
  CommentOutlined,
  DashboardOutlined,
} from '@vicons/material';
import { AdminKitApp, AdminKitLayout } from '@novelia/admin-kit';
import { NIcon, type MenuOption } from 'naive-ui';
import { h, type Component } from 'vue';
import { RouterView, useRoute } from 'vue-router';

const route = useRoute();

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) });
}

const menuOptions: MenuOption[] = [
  { label: '概览', key: '/overview', icon: renderIcon(DashboardOutlined) },
  { label: '分类管理', key: '/categories', icon: renderIcon(CategoryOutlined) },
  { label: '帖子管理', key: '/posts', icon: renderIcon(ArticleOutlined) },
  { label: '评论审核', key: '/comments', icon: renderIcon(CommentOutlined) },
];
</script>

<template>
  <AdminKitApp>
    <AdminKitLayout
      v-if="route.meta.requiresAuth"
      :menu-options="menuOptions"
    />
    <RouterView v-else />
  </AdminKitApp>
</template>
