<script setup lang="ts">
import { NButton, NEmpty, NSkeleton, NTag } from 'naive-ui';

import type { Tag } from '@/api';

defineProps<{
  categoryName: string;
  tags: Tag[];
  loading: boolean;
  activeUpdatingId?: number;
}>();

const emit = defineEmits<{
  create: [];
  edit: [tag: Tag];
  toggleActive: [tag: Tag];
}>();
</script>

<template>
  <section class="tag-panel">
    <div class="panel-header">
      <h2>{{ categoryName }}</h2>
      <n-button type="primary" :disabled="loading" @click="emit('create')">
        新建标签
      </n-button>
    </div>

    <div v-if="loading" class="skeleton-stack">
      <n-skeleton v-for="index in 3" :key="index" text />
    </div>
    <n-empty v-else-if="!tags.length" description="暂无标签" />
    <ul v-else class="tag-list">
      <li v-for="tag in tags" :key="tag.id" class="tag-row">
        <div class="tag-copy">
          <div class="tag-title">
            <strong>{{ tag.name }}</strong>
            <n-tag :type="tag.isActive ? 'success' : 'default'" size="small">
              {{ tag.isActive ? '启用' : '停用' }}
            </n-tag>
          </div>
          <span class="tag-detail">
            色号 {{ tag.color }} · 排序 {{ tag.sortOrder }}
          </span>
        </div>
        <div class="tag-actions">
          <n-button size="small" quaternary @click="emit('edit', tag)">
            编辑
          </n-button>
          <n-button
            size="small"
            quaternary
            :loading="activeUpdatingId === tag.id"
            @click="emit('toggleActive', tag)"
          >
            {{ tag.isActive ? '停用' : '启用' }}
          </n-button>
        </div>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.panel-header h2 {
  margin: 0;
  font-size: 16px;
}

.tag-panel {
  padding: 20px;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
}

.tag-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.tag-copy {
  min-width: 0;
  flex: 1;
}

.tag-row {
  padding: 12px 0;
  border-bottom: 1px solid var(--n-border-color);
  display: flex;
  align-items: center;
  gap: 12px;
}

.tag-row:last-child {
  border-bottom: 0;
}

.tag-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-detail {
  color: var(--n-text-color-3);
  font-size: 12px;
}

.tag-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: none;
}

.skeleton-stack {
  display: grid;
  gap: 12px;
}

@media (max-width: 520px) {
  .tag-row {
    align-items: flex-start;
    flex-direction: column;
    gap: 4px;
  }
}
</style>
