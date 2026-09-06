<script setup lang="ts">
import { AdminKitApp, AdminKitLayout } from '@novelia/admin-kit';
import type { MenuOption } from 'naive-ui';
import { RouterView, useRoute } from 'vue-router';

const route = useRoute();
const menuOptions: MenuOption[] = [];
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
