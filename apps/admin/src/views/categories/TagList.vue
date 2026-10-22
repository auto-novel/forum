<script setup lang="ts">
import { XButton } from '@novelia/web-kit';

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
  <section class="py-6">
    <div class="mb-3 flex items-center justify-between gap-3">
      <h2 class="m-0 text-base font-semibold text-ink">{{ categoryName }}</h2>
      <XButton :disabled="loading" @click="emit('create')">新建标签</XButton>
    </div>

    <div v-if="loading" class="grid gap-3">
      <div
        v-for="index in 3"
        :key="index"
        aria-hidden="true"
        class="skeleton h-4"
      ></div>
    </div>
    <p v-else-if="!tags.length" class="py-8 text-center text-sm text-muted">
      暂无标签
    </p>
    <ul v-else class="divide-y divide-divider">
      <li
        v-for="tag in tags"
        :key="tag.id"
        class="flex flex-col items-start gap-1 py-3 sm:flex-row sm:items-center sm:gap-3"
      >
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <strong class="text-ink">{{ tag.name }}</strong>
            <span
              class="inline-flex rounded px-2 py-0.5 text-xs"
              :class="
                tag.isActive
                  ? 'bg-success-soft text-success-text'
                  : 'bg-ink/5 text-muted'
              "
            >
              {{ tag.isActive ? '启用' : '停用' }}
            </span>
          </div>
          <span class="text-xs text-muted">
            色号 {{ tag.color }} · 排序 {{ tag.sortOrder }}
          </span>
        </div>
        <div class="flex shrink-0 items-center gap-1">
          <XButton
            size="sm"
            color="neutral"
            variant="ghost"
            @click="emit('edit', tag)"
          >
            编辑
          </XButton>
          <XButton
            size="sm"
            color="neutral"
            variant="ghost"
            :loading="activeUpdatingId === tag.id"
            @click="emit('toggleActive', tag)"
          >
            {{ tag.isActive ? '停用' : '启用' }}
          </XButton>
        </div>
      </li>
    </ul>
  </section>
</template>
