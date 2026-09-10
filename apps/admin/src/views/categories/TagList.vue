<script setup lang="ts">
import { AddOutlined, LocalOfferOutlined } from '@vicons/material';
import {
  NButton,
  NCard,
  NEmpty,
  NIcon,
  NSkeleton,
  NTag,
  NText,
} from 'naive-ui';

import type { Category, Tag } from '@/api';

defineProps<{
  category?: Category;
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
  <n-card class="tag-panel" :bordered="false">
    <template #header>
      <div class="panel-header">
        <div class="panel-title">
          <n-text strong>{{ category?.slug ?? '标签' }}</n-text>
          <n-text depth="3" class="panel-caption">
            {{ tags.length }} 个标签
          </n-text>
        </div>
        <n-button
          size="small"
          secondary
          type="primary"
          :disabled="!category"
          @click="emit('create')"
        >
          <template #icon><n-icon :component="AddOutlined" /></template>
          新建标签
        </n-button>
      </div>
    </template>

    <div v-if="loading" class="skeleton-stack">
      <n-skeleton v-for="index in 4" :key="index" text :repeat="2" />
    </div>
    <n-empty
      v-else-if="!category || !tags.length"
      :description="category ? '该分类暂无标签' : '请先创建分类'"
    />
    <div v-else class="tag-list">
      <div v-for="tag in tags" :key="tag.id" class="tag-row">
        <span class="tag-symbol">
          <n-icon :component="LocalOfferOutlined" />
        </span>
        <div class="tag-copy">
          <div class="tag-title">
            <n-text strong>{{ tag.name }}</n-text>
            <n-tag :type="tag.isActive ? 'success' : 'default'" size="small">
              {{ tag.isActive ? '启用' : '停用' }}
            </n-tag>
          </div>
          <n-text depth="3">
            色号 {{ tag.color }} · 排序 {{ tag.sortOrder }}
          </n-text>
        </div>
        <div class="tag-actions">
          <n-button
            size="small"
            quaternary
            :type="tag.isActive ? 'error' : 'success'"
            :loading="activeUpdatingId === tag.id"
            @click="emit('toggleActive', tag)"
          >
            {{ tag.isActive ? '停用' : '启用' }}
          </n-button>
          <n-button size="small" quaternary @click="emit('edit', tag)">
            编辑
          </n-button>
        </div>
      </div>
    </div>
  </n-card>
</template>

<style scoped>
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.panel-title,
.tag-copy {
  min-width: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 3px;
}

.panel-caption {
  font-size: 12px;
}

.tag-panel {
  min-height: 360px;
}

.tag-list {
  display: grid;
  gap: 2px;
}

.tag-row {
  min-height: 64px;
  padding: 10px 4px;
  border-bottom: 1px solid var(--n-border-color);
  display: flex;
  align-items: center;
  gap: 12px;
}

.tag-row:last-child {
  border-bottom: 0;
}

.tag-symbol {
  width: 34px;
  height: 34px;
  border-radius: 9px;
  color: #2080f0;
  background: rgba(32, 128, 240, 0.1);
  display: grid;
  place-items: center;
  flex: none;
}

.tag-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.skeleton-stack {
  display: grid;
  gap: 12px;
}
</style>
