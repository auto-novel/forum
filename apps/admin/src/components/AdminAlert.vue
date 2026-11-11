<script setup lang="ts">
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
const icons = {
  info: 'i-lucide-info',
  success: 'i-lucide-circle-check',
  warning: 'i-lucide-triangle-alert',
  error: 'i-lucide-circle-alert',
};
</script>

<template>
  <UAlert
    role="alert"
    :color="type"
    variant="soft"
    :title="title"
    :icon="icons[type]"
    :close="closable ? { size: 'xl' } : false"
    :ui="{ description: 'break-words' }"
    @update:open="(open) => !open && emit('close')"
  >
    <template v-if="$slots.default" #description>
      <slot />
    </template>
  </UAlert>
</template>
