<script setup lang="ts">
import {
  CheckCircleOutlineOutlined,
  CloseOutlined,
  ErrorOutlineOutlined,
  InfoOutlined,
  WarningAmberOutlined,
} from '@vicons/material';

withDefaults(
  defineProps<{
    type?: 'info' | 'success' | 'warning' | 'error';
    title?: string;
    closable?: boolean;
  }>(),
  {
    type: 'info',
    closable: false,
  },
);

const emit = defineEmits<{ close: [] }>();

const tones = {
  info: 'bg-info-soft text-info-text',
  success: 'bg-success-soft text-success-text',
  warning: 'bg-warning-soft text-warning-text',
  error: 'bg-error-soft text-error-text',
};
const icons = {
  info: InfoOutlined,
  success: CheckCircleOutlineOutlined,
  warning: WarningAmberOutlined,
  error: ErrorOutlineOutlined,
};
</script>

<template>
  <div
    role="alert"
    class="flex items-start gap-3 rounded-lg p-4 text-sm"
    :class="tones[type]"
  >
    <component
      :is="icons[type]"
      class="mt-0.5 size-5 shrink-0"
      aria-hidden="true"
    />
    <div class="min-w-0 flex-1 break-words">
      <p v-if="title" class="font-medium" :class="{ 'mb-2': $slots.default }">
        {{ title }}
      </p>
      <slot />
    </div>
    <button
      v-if="closable"
      type="button"
      aria-label="关闭"
      class="shrink-0 cursor-pointer rounded p-0.5 hover:bg-neutral-soft focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-current"
      @click="emit('close')"
    >
      <CloseOutlined class="size-4" aria-hidden="true" />
    </button>
  </div>
</template>
