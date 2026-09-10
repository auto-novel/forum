<script setup lang="ts">
import {
  DarkModeOutlined,
  LightModeOutlined,
  SmartToyOutlined,
} from '@vicons/material';

import type { Category } from '@/api';
import { useTheme } from '@/theme';

import CategoryNavigation from './CategoryNavigation.vue';

defineProps<{
  categories: Category[];
  selected?: string;
  collapsed?: boolean;
  fullWidth?: boolean;
}>();

const emit = defineEmits<{
  select: [slug: string];
}>();

const { isDark, toggleTheme } = useTheme();
</script>

<template>
  <aside
    class="forum-sidebar flex h-full flex-col overflow-hidden border-r border-divider bg-surface"
    :class="fullWidth ? 'w-full' : collapsed ? 'w-16' : 'w-56'"
    aria-label="论坛导航"
  >
    <div class="flex h-16 min-w-56 flex-none items-center px-4">
      <span
        class="grid size-8 place-items-center rounded-md bg-primary-soft text-primary"
        aria-hidden="true"
      >
        <SmartToyOutlined class="size-5" />
      </span>
      <span
        class="sidebar-label ml-2.5 text-sm font-bold tracking-tight whitespace-nowrap text-ink"
        :class="collapsed ? 'opacity-0' : 'opacity-100'"
        :aria-hidden="collapsed"
      >
        论坛
      </span>
    </div>

    <div class="min-h-0 flex-1 overflow-x-hidden overflow-y-auto p-2">
      <CategoryNavigation
        :categories="categories"
        :selected="selected"
        :collapsed="collapsed"
        @select="emit('select', $event)"
      />
    </div>

    <div class="flex-none p-2">
      <button
        type="button"
        class="theme-toggle"
        :aria-label="isDark ? '切换到浅色主题' : '切换到深色主题'"
        :title="isDark ? '切换到浅色主题' : '切换到深色主题'"
        @click="toggleTheme"
      >
        <span
          class="grid size-7 flex-none place-items-center"
          aria-hidden="true"
        >
          <LightModeOutlined v-if="isDark" class="size-5" />
          <DarkModeOutlined v-else class="size-5" />
        </span>
        <span
          class="sidebar-label whitespace-nowrap"
          :class="collapsed ? 'opacity-0' : 'opacity-100'"
          :aria-hidden="collapsed"
        >
          {{ isDark ? '浅色主题' : '深色主题' }}
        </span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.forum-sidebar {
  transition: width 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar-label {
  transition: opacity 150ms ease;
}

.theme-toggle {
  display: flex;
  min-height: 2.75rem;
  width: 100%;
  min-width: 13rem;
  align-items: center;
  gap: 0.7rem;
  border-radius: 0.25rem;
  padding-inline: 0.625rem;
  color: var(--color-muted);
  font-size: 0.875rem;
  font-weight: 600;
  text-align: left;
  transition:
    color 150ms ease,
    background-color 150ms ease;
}

.theme-toggle:hover {
  background: var(--color-paper);
  color: var(--color-ink);
}

.theme-toggle:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
</style>
