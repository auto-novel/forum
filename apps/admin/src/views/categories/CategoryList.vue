<script setup lang="ts">
import { EditOutlined, ImageOutlined } from '@vicons/material';
import { NButton, NEmpty, NIcon, NSkeleton, NText } from 'naive-ui';

import type { Category } from '@/api';

defineProps<{
  categories: Category[];
  loading: boolean;
  selectedId?: number;
}>();

const emit = defineEmits<{
  select: [id: number];
  edit: [category: Category];
}>();
</script>

<template>
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
      :class="['category-card', { selected: selectedId === category.id }]"
      @click="emit('select', category.id)"
    >
      <span class="category-art">
        <img v-if="category.bannerUrl" :src="category.bannerUrl" alt="" />
        <n-icon v-else :component="ImageOutlined" />
      </span>
      <span class="category-copy">
        <n-text strong>{{ category.slug }}</n-text>
        <n-text depth="3">ID {{ category.id }}</n-text>
      </span>
      <n-button text aria-label="编辑分类" @click.stop="emit('edit', category)">
        <template #icon><n-icon :component="EditOutlined" /></template>
      </n-button>
    </button>
  </section>
</template>

<style scoped>
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

.category-copy {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 3px;
}

.skeleton-stack {
  display: grid;
  gap: 12px;
}
</style>
