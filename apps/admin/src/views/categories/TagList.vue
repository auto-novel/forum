<script setup lang="ts">
import { XSkeleton } from '@novelia/web-kit';
import { XButton } from '@novelia/web-kit';
import { AddOutlined, LocalOfferOutlined } from '@vicons/material';

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
  <section
    class="py-6"
    :aria-label="`${categoryName}标签`"
    :aria-busy="loading"
  >
    <header class="mb-4 flex items-center justify-between gap-4">
      <div class="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-1">
        <h2 class="text-base font-semibold text-ink">{{ categoryName }}</h2>
        <span v-if="!loading" class="text-xs tabular-nums text-muted">
          {{ tags.length }} 个标签
        </span>
      </div>
      <XButton
        size="sm"
        variant="outline"
        class="shrink-0"
        :disabled="loading"
        @click="emit('create')"
      >
        <AddOutlined class="size-4" aria-hidden="true" />
        新建标签
      </XButton>
    </header>

    <div v-if="loading">
      <p role="status" class="sr-only">正在加载{{ categoryName }}的标签</p>
      <div
        aria-hidden="true"
        class="divide-y divide-divider border-y border-divider"
      >
        <div
          v-for="index in 3"
          :key="index"
          class="flex items-center gap-6 px-3 py-4"
        >
          <XSkeleton class="h-4 w-1/3" />
          <XSkeleton class="h-4 w-12" />
          <XSkeleton class="ml-auto h-6 w-24" />
        </div>
      </div>
    </div>
    <div
      v-else-if="!tags.length"
      class="flex flex-col items-center border-y border-dashed border-divider px-4 py-8 text-center"
    >
      <LocalOfferOutlined class="mb-3 size-6 text-muted" aria-hidden="true" />
      <p class="text-sm text-ink">暂无标签</p>
      <p class="mt-1 text-xs text-muted">
        点击「新建标签」，为此分类添加标签。
      </p>
    </div>
    <div v-else>
      <div
        aria-hidden="true"
        class="hidden grid-cols-12 gap-4 border-b border-divider px-3 pb-2 text-xs text-muted sm:grid"
      >
        <span class="col-span-4">标签名称</span>
        <span class="col-span-2">状态</span>
        <span class="col-span-3">属性</span>
        <span class="col-span-3 text-right">操作</span>
      </div>
      <ul class="divide-y divide-divider border-b border-divider">
        <li
          v-for="tag in tags"
          :key="tag.id"
          class="grid grid-cols-12 items-center gap-x-4 gap-y-2 px-3 py-3 transition-colors hover:bg-hover focus-within:bg-hover"
        >
          <span
            class="col-span-8 min-w-0 break-words text-sm font-medium text-ink sm:col-span-4"
          >
            {{ tag.name }}
          </span>
          <span
            class="col-span-4 inline-flex items-center justify-end gap-1.5 text-xs sm:col-span-2 sm:justify-start"
            :class="tag.isActive ? 'text-success-text' : 'text-muted'"
          >
            <span
              class="size-1.5 shrink-0 rounded-full"
              :class="tag.isActive ? 'bg-success' : 'bg-muted'"
              aria-hidden="true"
            ></span>
            {{ tag.isActive ? '启用' : '停用' }}
          </span>
          <div
            class="col-span-6 flex flex-wrap gap-x-3 gap-y-1 text-xs tabular-nums text-muted sm:col-span-3"
          >
            <span>
              色号
              <span class="text-ink">{{ tag.color }}</span>
            </span>
            <span>
              排序
              <span class="text-ink">{{ tag.sortOrder }}</span>
            </span>
          </div>
          <div
            class="col-span-6 flex items-center justify-end gap-1 sm:col-span-3"
          >
            <XButton
              size="sm"
              color="neutral"
              variant="ghost"
              :aria-label="`编辑标签「${tag.name}」`"
              @click="emit('edit', tag)"
            >
              编辑
            </XButton>
            <XButton
              size="sm"
              :color="tag.isActive ? 'warning' : 'primary'"
              variant="ghost"
              :aria-label="`${tag.isActive ? '停用' : '启用'}标签「${tag.name}」`"
              :loading="activeUpdatingId === tag.id"
              @click="emit('toggleActive', tag)"
            >
              {{ tag.isActive ? '停用' : '启用' }}
            </XButton>
          </div>
        </li>
      </ul>
    </div>
  </section>
</template>
