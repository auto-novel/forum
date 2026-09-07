<script setup lang="ts">
import type { Component } from 'vue';
import { NCard, NIcon, NSkeleton, NText } from 'naive-ui';

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
</script>

<template>
  <section class="metric-grid" aria-label="论坛数据概览">
    <n-card v-for="metric in metrics" :key="metric.label" class="metric-card">
      <n-skeleton v-if="loading" text :repeat="2" />
      <div v-else class="metric-content">
        <div class="metric-copy">
          <n-text depth="3">{{ metric.label }}</n-text>
          <n-text class="metric-value">
            {{ metric.value.toLocaleString() }}
          </n-text>
        </div>
        <span :class="['metric-icon', metric.tone]">
          <n-icon :component="metric.icon" />
        </span>
      </div>
    </n-card>
  </section>
</template>

<style scoped>
.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metric-card :deep(.n-card__content) {
  min-height: 108px;
  padding: 18px;
}

.metric-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.metric-copy {
  display: flex;
  flex-direction: column;
  gap: 7px;
}

.metric-value {
  font-size: 28px;
  line-height: 1;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.metric-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  flex: none;
  font-size: 21px;
}

.metric-icon.blue {
  color: #2080f0;
  background: rgba(32, 128, 240, 0.12);
}

.metric-icon.green {
  color: #18a058;
  background: rgba(24, 160, 88, 0.12);
}

.metric-icon.orange {
  color: #f0a020;
  background: rgba(240, 160, 32, 0.14);
}

.metric-icon.purple {
  color: #8b5cf6;
  background: rgba(139, 92, 246, 0.12);
}

@media (max-width: 900px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .metric-grid {
    gap: 10px;
  }
}
</style>
