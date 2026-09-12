<script setup lang="ts">
import { ErrorOutlineOutlined, Inventory2Outlined } from '@vicons/material';
import { computed } from 'vue';

const props = withDefaults(
  defineProps<{
    loading: boolean;
    error?: string;
    empty?: boolean;
    errorTitle?: string;
    retryLabel?: string;
    emptyTitle?: string;
    emptyDescription?: string;
    size?: 'compact' | 'default' | 'large';
    headingTag?: 'h1' | 'h2' | 'p';
    stateClass?: string;
  }>(),
  {
    error: '',
    empty: false,
    errorTitle: '加载失败',
    retryLabel: '再试一次',
    emptyTitle: '暂无内容',
    size: 'default',
    headingTag: 'h2',
    stateClass: '',
  },
);

defineEmits<{ retry: [] }>();

const minHeightClass = computed(
  () =>
    ({
      compact: 'min-h-52',
      default: 'min-h-80',
      large: 'min-h-96',
    })[props.size],
);
</script>

<template>
  <div class="contents">
    <slot v-if="loading" name="loading" />

    <div
      v-else-if="error"
      class="grid place-items-center p-8 text-center"
      :class="[minHeightClass, stateClass]"
    >
      <div>
        <slot name="error-icon">
          <div
            class="mx-auto grid size-12 place-items-center rounded-full bg-red-50 text-red-500"
            aria-hidden="true"
          >
            <ErrorOutlineOutlined class="size-6" />
          </div>
        </slot>
        <component :is="headingTag" class="mt-4 text-lg font-semibold text-ink">
          {{ errorTitle }}
        </component>
        <p class="mt-2 text-sm text-muted">{{ error }}</p>
        <button
          type="button"
          class="mt-5 rounded-sm bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
          @click="$emit('retry')"
        >
          {{ retryLabel }}
        </button>
      </div>
    </div>

    <div
      v-else-if="empty"
      class="grid place-items-center p-8 text-center"
      :class="[minHeightClass, stateClass]"
    >
      <div>
        <slot name="empty-icon">
          <div
            class="mx-auto grid size-12 place-items-center rounded-full bg-paper text-muted"
            aria-hidden="true"
          >
            <Inventory2Outlined class="size-6" />
          </div>
        </slot>
        <component :is="headingTag" class="mt-4 text-lg font-semibold text-ink">
          {{ emptyTitle }}
        </component>
        <p v-if="emptyDescription" class="mt-2 text-sm text-muted">
          {{ emptyDescription }}
        </p>
      </div>
    </div>

    <slot v-else />
  </div>
</template>
