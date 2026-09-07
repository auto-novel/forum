<script setup lang="ts">
import { NFlex, NText } from 'naive-ui';

interface FilterChoiceOption {
  label: string;
  value: string;
}

defineProps<{
  value: string;
  options: FilterChoiceOption[];
}>();

const emit = defineEmits<{
  'update:value': [value: string];
}>();

function selectWithKeyboard(event: KeyboardEvent, value: string) {
  if (event.key !== 'Enter' && event.key !== ' ') return;
  event.preventDefault();
  emit('update:value', value);
}
</script>

<template>
  <n-flex :size="[16, 4]" role="radiogroup">
    <n-text
      v-for="option in options"
      :key="option.value"
      class="filter-choice"
      :class="{ 'filter-choice--selected': value === option.value }"
      :type="value === option.value ? 'primary' : 'default'"
      role="radio"
      :aria-checked="value === option.value"
      tabindex="0"
      @click="emit('update:value', option.value)"
      @keydown="selectWithKeyboard($event, option.value)"
    >
      {{ option.label }}
    </n-text>
  </n-flex>
</template>

<style scoped>
.filter-choice {
  cursor: pointer;
  user-select: none;
}

.filter-choice:focus-visible {
  border-radius: 2px;
  outline: 2px solid currentColor;
  outline-offset: 2px;
}

.filter-choice--selected {
  font-weight: 500;
}
</style>
