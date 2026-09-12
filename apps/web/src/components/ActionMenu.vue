<script setup lang="ts">
import { MoreHorizOutlined } from '@vicons/material';
import {
  DropdownMenuContent,
  DropdownMenuPortal,
  DropdownMenuRoot,
  DropdownMenuTrigger,
} from 'reka-ui';

withDefaults(
  defineProps<{
    side?: 'top' | 'right' | 'bottom' | 'left';
    align?: 'start' | 'center' | 'end';
    compact?: boolean;
  }>(),
  {
    side: 'bottom',
    align: 'end',
  },
);
</script>

<template>
  <DropdownMenuRoot>
    <DropdownMenuTrigger as-child>
      <button
        type="button"
        class="inline-flex items-center justify-center rounded-sm text-muted transition-colors hover:bg-paper hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary"
        :class="compact ? 'size-[1.875rem]' : 'size-9'"
        aria-label="更多操作"
        title="更多操作"
      >
        <MoreHorizOutlined
          :class="compact ? 'size-4' : 'size-5'"
          aria-hidden="true"
        />
      </button>
    </DropdownMenuTrigger>
    <DropdownMenuPortal>
      <DropdownMenuContent
        :side="side"
        :align="align"
        :side-offset="6"
        :collision-padding="8"
        class="z-30 rounded-md border border-border bg-surface p-1 shadow-xl outline-none"
        :class="compact ? 'w-32' : 'w-40'"
      >
        <slot />
      </DropdownMenuContent>
    </DropdownMenuPortal>
  </DropdownMenuRoot>
</template>
