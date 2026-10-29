<script setup lang="ts">
import { XButton } from '@novelia/web-kit';
import { VerticalAlignTopOutlined } from '@vicons/material';
import { onMounted, onScopeDispose, ref } from 'vue';

const visible = ref(false);

function updateVisibility() {
  visible.value = window.scrollY > 1200;
}

function backToTop() {
  window.scrollTo({
    top: 0,
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches
      ? 'instant'
      : 'smooth',
  });
}

onMounted(() => {
  updateVisibility();
  window.addEventListener('scroll', updateVisibility, { passive: true });
});
onScopeDispose(() => {
  window.removeEventListener('scroll', updateVisibility);
});
</script>

<template>
  <XButton
    v-if="visible"
    variant="soft"
    class="fixed right-4 bottom-[calc(1rem+env(safe-area-inset-bottom))] z-30 size-11 rounded-full shadow-lg md:right-8 md:bottom-8"
    title="回到顶部"
    aria-label="回到顶部"
    @click="backToTop"
  >
    <VerticalAlignTopOutlined class="size-5 shrink-0" aria-hidden="true" />
  </XButton>
</template>
