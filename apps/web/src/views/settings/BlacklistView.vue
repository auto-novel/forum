<script setup lang="ts">
import { BlockOutlined } from '@vicons/material';
import {
  XError,
  XEmpty,
  XSkeleton,
  XButton,
  XConfirmDialog,
  XTime,
} from '@novelia/web-kit';
import { useBlacklistManagement } from './useBlacklistManagement';

const {
  whoami,
  entries,
  loading,
  error,
  selected,
  removing,
  load,
  confirmRemove,
} = useBlacklistManagement();
</script>

<template>
  <section aria-labelledby="blacklist-heading">
    <header class="mb-4 border-b border-divider pb-3">
      <h2
        id="blacklist-heading"
        class="text-2xl font-normal tracking-tight text-ink"
      >
        黑名单管理
      </h2>
    </header>
    <p class="mb-5 text-sm leading-relaxed text-muted">
      拉黑后，帖子列表和收藏列表将不再显示该用户的帖子，评论区将隐藏其评论及其发起的评论串。
    </p>

    <div v-if="!whoami.isSignedIn" class="py-8 text-center">
      <div
        class="mx-auto grid size-12 place-items-center rounded-full bg-primary-soft text-primary"
        aria-hidden="true"
      >
        <BlockOutlined class="size-6" />
      </div>
      <h2 class="mt-4 text-lg font-semibold text-ink">登录后管理黑名单</h2>
      <p class="mt-2 text-sm text-muted">
        请使用页面右上角的登录入口登录账号。
      </p>
    </div>

    <div v-else aria-live="polite">
      <div v-if="loading" role="status">
        <span class="sr-only">正在加载黑名单…</span>
        <div aria-hidden="true">
          <XSkeleton class="mb-3 h-5 w-32" />
          <div
            class="divide-y divide-divider rounded-md border border-divider px-4"
          >
            <div
              v-for="index in 3"
              :key="index"
              class="flex items-center justify-between gap-4 py-4"
            >
              <div class="min-w-0 flex-1">
                <XSkeleton class="h-6 w-32 max-w-full bg-border" />
                <XSkeleton class="mt-1 h-4 w-40 max-w-full" />
              </div>
              <XSkeleton class="h-9 w-22 shrink-0" />
            </div>
          </div>
        </div>
      </div>
      <XError v-else-if="error" title="黑名单加载失败" :description="error">
        <XButton @click="load">再试一次</XButton>
      </XError>
      <XEmpty
        v-else-if="!entries.length"
        title="黑名单为空"
        description="你还没有拉黑任何用户。"
      />
      <template v-else>
        <p class="mb-3 text-sm text-muted">
          已拉黑 {{ entries.length }} 位用户
        </p>
        <ul
          class="divide-y divide-divider rounded-md border border-divider px-4"
        >
          <li
            v-for="entry in entries"
            :key="entry.userId"
            class="flex items-center justify-between gap-4 py-4"
          >
            <div class="min-w-0">
              <p class="break-words font-medium text-ink">
                用户 ID：{{ entry.userId }}
              </p>
              <p class="mt-1 text-xs text-muted">
                拉黑于
                <XTime :time="entry.createdAt" preset="date" />
              </p>
            </div>
            <XButton
              color="neutral"
              variant="outline"
              size="sm"
              class="shrink-0"
              :disabled="removing"
              :aria-label="`取消拉黑用户 ${entry.userId}`"
              @click="selected = entry"
            >
              取消拉黑
            </XButton>
          </li>
        </ul>
      </template>
    </div>

    <XConfirmDialog
      v-if="selected"
      open
      :title="`取消拉黑用户 ${selected.userId}`"
      description="取消后，该用户的帖子和评论将重新出现在列表中。"
      confirm-label="取消拉黑"
      :loading="removing"
      @update:open="
        (open) => {
          if (!open && !removing) selected = undefined;
        }
      "
      @confirm="confirmRemove"
    />
  </section>
</template>
