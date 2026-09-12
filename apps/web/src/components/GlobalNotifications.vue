<script setup lang="ts">
import {
  CheckCircleOutlineOutlined,
  CloseOutlined,
  ErrorOutlineOutlined,
} from '@vicons/material';
import {
  ToastClose,
  ToastDescription,
  ToastPortal,
  ToastProvider,
  ToastRoot,
  ToastTitle,
  ToastViewport,
} from 'reka-ui';

import { dismissNotification, notifications } from '@/notifications';
</script>

<template>
  <ToastProvider
    label="通知"
    :duration="4000"
    swipe-direction="up"
    :swipe-threshold="40"
  >
    <ToastRoot
      v-for="notification in notifications"
      :key="notification.id"
      class="flex w-full items-start gap-3 rounded-md border bg-surface px-4 py-3 shadow-xl outline-none data-[swipe=cancel]:translate-y-0 data-[swipe=end]:translate-y-[var(--reka-toast-swipe-end-y)] data-[swipe=move]:translate-y-[var(--reka-toast-swipe-move-y)] data-[swipe=cancel]:transition-transform data-[swipe=end]:transition-transform"
      :class="
        notification.type === 'success' ? 'border-green-200' : 'border-red-200'
      "
      @update:open="
        (open) => {
          if (!open) dismissNotification(notification.id);
        }
      "
    >
      <CheckCircleOutlineOutlined
        v-if="notification.type === 'success'"
        class="mt-0.5 size-5 flex-none text-green-600"
        aria-hidden="true"
      />
      <ErrorOutlineOutlined
        v-else
        class="mt-0.5 size-5 flex-none text-red-600"
        aria-hidden="true"
      />
      <div class="min-w-0 flex-1">
        <ToastTitle class="text-sm font-semibold text-ink">
          {{ notification.title }}
        </ToastTitle>
        <ToastDescription class="mt-0.5 text-sm leading-5 text-muted">
          {{ notification.description }}
        </ToastDescription>
      </div>
      <ToastClose
        class="grid size-7 flex-none place-items-center rounded-sm text-muted transition-colors hover:bg-paper hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary"
        aria-label="关闭通知"
      >
        <CloseOutlined class="size-4" aria-hidden="true" />
      </ToastClose>
    </ToastRoot>

    <ToastPortal>
      <ToastViewport
        class="fixed top-4 left-1/2 z-[100] flex max-h-[calc(100vh-2rem)] w-[calc(100%-2rem)] max-w-sm -translate-x-1/2 flex-col gap-2 outline-none"
        label="通知（{hotkey}）"
      />
    </ToastPortal>
  </ToastProvider>
</template>
