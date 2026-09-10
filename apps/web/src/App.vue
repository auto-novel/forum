<script setup lang="ts">
import { RouterView } from 'vue-router';

import UserAccountButton from '@/components/UserAccountButton.vue';
</script>

<template>
  <div class="min-h-screen">
    <header class="sticky top-0 z-30 border-b border-divider bg-surface">
      <div class="page-container flex h-16 items-center">
        <UserAccountButton />
      </div>
    </header>

    <main>
      <RouterView />
    </main>
  </div>
</template>
