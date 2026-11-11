<script setup lang="ts">
import { useId } from 'vue';

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
const groupName = useId();

function selectFocused(event: KeyboardEvent) {
  if (
    event.target instanceof HTMLElement &&
    event.target.getAttribute('role') === 'radio'
  ) {
    event.preventDefault();
    event.target.click();
  }
}
</script>

<template>
  <URadioGroup
    :name="groupName"
    :model-value="value"
    :items="options"
    orientation="horizontal"
    :ui="{
      fieldset: 'gap-x-4',
      item: 'rounded-sm has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-primary',
      container: 'sr-only',
      label: 'cursor-pointer',
    }"
    @update:model-value="emit('update:value', String($event))"
    @keydown.enter="selectFocused"
  >
    <template #label="{ item }">
      <span
        :class="value === item.value ? 'font-medium text-primary' : 'text-ink'"
      >
        {{ item.label }}
      </span>
    </template>
  </URadioGroup>
</template>
