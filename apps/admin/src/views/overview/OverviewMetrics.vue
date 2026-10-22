<script setup lang="ts">
import type { Component } from 'vue';

interface Metric {
  label: string;
  value: number;
  icon: Component;
  tone: string;
}

defineProps<{
  metrics: Metric[];
  loading: boolean;
}>();

const tones: Record<string, string> = {
  info: 'bg-info-soft text-info',
  success: 'bg-success-soft text-success',
  warning: 'bg-warning-soft text-warning',
  primary: 'bg-primary-soft text-primary',
};
</script>

<template>
  <section
    class="grid grid-cols-4 gap-3.5 max-[900px]:grid-cols-2 max-[640px]:gap-2.5"
    aria-label="论坛数据概览"
    :aria-busy="loading"
  >
    <div
      v-for="metric in metrics"
      :key="metric.label"
      class="min-h-[108px] rounded-lg border border-border bg-surface p-[18px] text-ink"
    >
      <div v-if="loading" aria-hidden="true" class="grid gap-2">
        <div class="skeleton h-3.5"></div>
        <div class="skeleton h-3.5 w-3/5"></div>
      </div>
      <div v-else class="flex items-center justify-between gap-3">
        <div class="flex min-w-0 flex-col gap-[7px]">
          <span class="text-muted">{{ metric.label }}</span>
          <span class="text-[28px] leading-none font-bold tabular-nums">
            {{ metric.value.toLocaleString() }}
          </span>
        </div>
        <span
          class="grid size-[42px] shrink-0 place-items-center rounded-xl"
          :class="tones[metric.tone]"
        >
          <component :is="metric.icon" class="size-[21px]" aria-hidden="true" />
        </span>
      </div>
    </div>
  </section>
</template>
