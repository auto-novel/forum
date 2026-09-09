<script setup lang="ts">
import { inject, nextTick, onBeforeUnmount, onMounted, watch } from 'vue';

import type { Category } from '@/api';

import CategoryNavigation from './CategoryNavigation.vue';
import { postListMobileNavigationKey } from './mobileNavigation';

defineProps<{
  categories: Category[];
  selected: string;
}>();

const emit = defineEmits<{
  select: [slug: string];
}>();

const mobileNavigation = inject(postListMobileNavigationKey);
const desktopMediaQuery = window.matchMedia('(min-width: 768px)');
let previousBodyOverflow = '';
let bodyScrollLocked = false;

function closeMobileNavigation(restoreFocus = false) {
  if (!mobileNavigation) return;
  mobileNavigation.open.value = false;
  if (restoreFocus) {
    void nextTick(() =>
      document
        .querySelector<HTMLElement>('#mobile-navigation-trigger')
        ?.focus(),
    );
  }
}

function selectCategory(slug: string) {
  emit('select', slug);
  closeMobileNavigation();
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && mobileNavigation?.open.value) {
    closeMobileNavigation(true);
  }
}

function handleViewportChange(event: MediaQueryListEvent) {
  if (event.matches) closeMobileNavigation();
}

watch(
  () => mobileNavigation?.open.value,
  (open) => {
    if (open) {
      previousBodyOverflow = document.body.style.overflow;
      document.body.style.overflow = 'hidden';
      bodyScrollLocked = true;
      void nextTick(() =>
        document
          .querySelector<HTMLElement>('#mobile-category-navigation button')
          ?.focus(),
      );
    } else if (bodyScrollLocked) {
      document.body.style.overflow = previousBodyOverflow;
      bodyScrollLocked = false;
    }
  },
);

onMounted(() => {
  window.addEventListener('keydown', handleKeydown);
  desktopMediaQuery.addEventListener('change', handleViewportChange);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown);
  desktopMediaQuery.removeEventListener('change', handleViewportChange);
  if (bodyScrollLocked) document.body.style.overflow = previousBodyOverflow;
});
</script>

<template>
  <aside class="sticky top-24 hidden self-start md:block" aria-label="帖子分类">
    <div class="rounded-sm bg-surface p-2">
      <CategoryNavigation
        :categories="categories"
        :selected="selected"
        collapsed
        @select="selectCategory"
      />
    </div>
  </aside>

  <Teleport to="body">
    <Transition name="mobile-drawer">
      <div
        v-if="mobileNavigation?.open.value"
        class="fixed inset-0 z-40 md:hidden"
        role="dialog"
        aria-modal="true"
        aria-label="分类导航"
      >
        <button
          type="button"
          class="mobile-drawer-backdrop absolute inset-0 bg-black/45"
          aria-label="关闭分类导航"
          @click="closeMobileNavigation(true)"
        />
        <aside
          id="mobile-category-navigation"
          class="mobile-drawer-panel absolute inset-y-0 left-0 w-[min(17.5rem,calc(100vw-3rem))] bg-surface shadow-2xl"
        >
          <div
            class="flex h-16 items-center justify-between border-b border-divider px-4"
          >
            <span class="font-semibold">帖子分类</span>
            <button
              type="button"
              class="grid size-9 place-items-center rounded-full transition-colors hover:bg-paper focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
              aria-label="关闭分类导航"
              @click="closeMobileNavigation(true)"
            >
              <svg
                viewBox="0 0 24 24"
                class="size-5"
                fill="none"
                aria-hidden="true"
              >
                <path
                  d="m7 7 10 10M17 7 7 17"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                />
              </svg>
            </button>
          </div>
          <div class="p-2">
            <CategoryNavigation
              :categories="categories"
              :selected="selected"
              @select="selectCategory"
            />
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.mobile-drawer-enter-active,
.mobile-drawer-leave-active,
.mobile-drawer-enter-active .mobile-drawer-backdrop,
.mobile-drawer-leave-active .mobile-drawer-backdrop,
.mobile-drawer-enter-active .mobile-drawer-panel,
.mobile-drawer-leave-active .mobile-drawer-panel {
  transition: 0.25s ease;
}

.mobile-drawer-enter-from .mobile-drawer-backdrop,
.mobile-drawer-leave-to .mobile-drawer-backdrop {
  opacity: 0;
}

.mobile-drawer-enter-from .mobile-drawer-panel,
.mobile-drawer-leave-to .mobile-drawer-panel {
  transform: translateX(-100%);
}
</style>
