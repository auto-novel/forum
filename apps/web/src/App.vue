<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router';

import UserAccountButton from '@/components/UserAccountButton.vue';
</script>

<template>
  <div class="min-h-screen">
    <header class="sticky top-0 z-30 border-b border-divider bg-surface">
      <div class="page-container flex h-16 items-center">
        <RouterLink
          to="/posts"
          class="flex items-center gap-2.5 rounded-sm focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
          aria-label="Novelia Forum 首页"
        >
          <span
            class="grid size-8 place-items-center rounded-md bg-primary-soft text-primary"
          >
            <svg
              viewBox="0 0 24 24"
              class="size-5"
              fill="none"
              aria-hidden="true"
            >
              <path
                d="M6.5 5.5h11v9h-6L7 18v-3.5h-.5v-9Z"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linejoin="round"
              />
              <path
                d="M9 9h6M9 12h3.5"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
              />
            </svg>
          </span>
          <span class="flex items-center gap-1.5 text-base leading-none">
            <span class="font-extrabold tracking-[-0.055em]">Forum</span>
            <span
              class="hidden size-1 rotate-45 rounded-[1px] bg-primary sm:block"
              aria-hidden="true"
            />
            <span
              class="hidden font-semibold tracking-tight text-primary sm:inline"
            >
              Community
            </span>
          </span>
        </RouterLink>

        <UserAccountButton />
      </div>
    </header>

    <main>
      <RouterView />
    </main>
  </div>
</template>
