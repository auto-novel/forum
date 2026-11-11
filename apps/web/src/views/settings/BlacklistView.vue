<script setup lang="ts">
import { BlockOutlined } from '@vicons/material';
import { XConfirmDialog, XTime } from '@novelia/web-kit';
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
  <section aria-label="黑名单管理">
    <UPageHeader
      title="黑名单管理"
      description="拉黑后，帖子列表和收藏列表将不再显示该用户的帖子，评论区将隐藏其评论及其发起的评论串。"
    />

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
          <USkeleton class="mb-3 h-5 w-32" />
          <div
            class="divide-y divide-divider rounded-md border border-divider px-4"
          >
            <div
              v-for="index in 3"
              :key="index"
              class="flex items-center justify-between gap-4 py-4"
            >
              <div class="min-w-0 flex-1">
                <USkeleton class="h-6 w-32 max-w-full" />
                <USkeleton class="mt-1 h-4 w-40 max-w-full" />
              </div>
              <USkeleton class="h-9 w-22 shrink-0" />
            </div>
          </div>
        </div>
      </div>
      <UError
        :ui="{
          root: 'min-h-0 py-8',
          statusMessage: 'text-lg font-medium sm:text-xl',
          message: 'mt-2 text-sm',
          links: 'mt-5 flex-wrap gap-3',
        }"
        v-else-if="error"
        :error="{ statusMessage: '黑名单加载失败', message: error }"
        :clear="false"
        :as="{ root: 'div', statusMessage: 'h2' }"
        role="alert"
      >
        <template #links>
          <UButton size="lg" @click="load">再试一次</UButton>
        </template>
      </UError>
      <UEmpty
        variant="naked"
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
            <UButton
              color="neutral"
              variant="outline"
              size="lg"
              class="shrink-0"
              :disabled="removing"
              :aria-label="`取消拉黑用户 ${entry.userId}`"
              @click="selected = entry"
            >
              取消拉黑
            </UButton>
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
