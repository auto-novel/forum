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
    class="flex h-full flex-col border-r border-divider bg-surface transition-[width] duration-300"
    :class="fullWidth ? 'w-full' : collapsed ? 'w-16' : 'w-56'"
    aria-label="论坛导航"
  >
    <div
      class="flex h-16 flex-none items-center"
      :class="collapsed ? 'justify-center px-2' : 'px-5'"
    >
      <span
        class="grid size-8 place-items-center rounded-md bg-primary-soft text-primary"
        aria-hidden="true"
      >
        <SmartToyOutlined class="size-5" />
      </span>
      <span
        v-if="!collapsed"
        class="ml-2.5 text-sm font-bold tracking-tight text-ink"
      >
        论坛
      </span>
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto p-2">
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
        :class="collapsed ? 'justify-center px-0' : 'px-3'"
        :aria-label="isDark ? '切换到浅色主题' : '切换到深色主题'"
        :title="isDark ? '切换到浅色主题' : '切换到深色主题'"
        @click="toggleTheme"
      >
        <LightModeOutlined v-if="isDark" class="size-5" aria-hidden="true" />
        <DarkModeOutlined v-else class="size-5" aria-hidden="true" />
        <span v-if="!collapsed">
          {{ isDark ? '浅色主题' : '深色主题' }}
        </span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.theme-toggle {
  display: flex;
  min-height: 2.75rem;
  width: 100%;
  align-items: center;
  gap: 0.7rem;
  border-radius: 0.25rem;
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
