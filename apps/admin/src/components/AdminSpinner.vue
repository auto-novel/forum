<script setup lang="ts">
withDefaults(defineProps<{ label?: string; size?: 'sm' | 'md' }>(), {
  label: '正在加载',
  size: 'md',
});
</script>

<template>
  <span
    class="inline-block animate-spin rounded-full border-2 border-current border-t-transparent text-muted motion-reduce:animate-none"
    :class="size === 'sm' ? 'size-4' : 'size-6'"
    role="status"
    :aria-label="label"
  />
</template>
