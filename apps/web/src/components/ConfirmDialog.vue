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

import AppButton from '@/components/AppButton.vue';

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
          <AlertDialogCancel as-child>
            <AppButton variant="outline" :disabled="loading">取消</AppButton>
          </AlertDialogCancel>
          <AlertDialogAction as-child>
            <AppButton
              :variant="danger ? 'danger' : 'primary'"
              :disabled="loading"
              @click="$emit('confirm')"
            >
              {{ loading ? '处理中…' : confirmLabel }}
            </AppButton>
          </AlertDialogAction>
        </div>
      </AlertDialogContent>
    </AlertDialogPortal>
  </AlertDialogRoot>
</template>
