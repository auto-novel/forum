<script setup lang="ts">
import {
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogOverlay,
  AlertDialogPortal,
  AlertDialogRoot,
  AlertDialogTitle,
} from 'reka-ui';

withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    description: string;
    confirmLabel?: string;
    loading?: boolean;
    danger?: boolean;
  }>(),
  {
    confirmLabel: '确认',
  },
);

defineEmits<{
  'update:open': [open: boolean];
  confirm: [];
}>();
</script>

<template>
  <AlertDialogRoot :open="open" @update:open="$emit('update:open', $event)">
    <AlertDialogPortal>
      <AlertDialogOverlay class="fixed inset-0 z-40 bg-black/45" />
      <AlertDialogContent
        class="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-md border border-border bg-surface p-5 shadow-2xl outline-none sm:p-6"
      >
        <AlertDialogTitle class="text-lg font-semibold text-ink">
          {{ title }}
        </AlertDialogTitle>
        <AlertDialogDescription class="mt-2 text-sm leading-6 text-muted">
          {{ description }}
        </AlertDialogDescription>
        <div class="mt-6 flex justify-end gap-3">
          <AlertDialogCancel
            class="rounded-sm border border-border px-4 py-2 text-sm font-medium text-muted transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="loading"
          >
            取消
          </AlertDialogCancel>
          <AlertDialogAction
            class="rounded-sm px-4 py-2 text-sm font-medium text-white transition-colors disabled:cursor-not-allowed disabled:opacity-50"
            :class="
              danger
                ? 'bg-red-600 hover:bg-red-700'
                : 'bg-primary hover:bg-primary-hover'
            "
            :disabled="loading"
            @click="$emit('confirm')"
          >
            {{ loading ? '处理中…' : confirmLabel }}
          </AlertDialogAction>
        </div>
      </AlertDialogContent>
    </AlertDialogPortal>
  </AlertDialogRoot>
</template>
