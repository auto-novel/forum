<script setup lang="ts">
import { RouterLink } from 'vue-router';
</script>

<template>
  <RouterLink
    :to="{ name: 'community-rules' }"
    class="block rounded-md bg-primary-soft px-3 py-2 text-sm font-medium text-primary hover:text-primary-hover"
  >
    发言请遵守社区守则
  </RouterLink>
</template>
