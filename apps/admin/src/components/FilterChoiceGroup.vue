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
</script>

<template>
  <div class="flex flex-wrap gap-x-4 gap-y-1" role="radiogroup">
    <label v-for="option in options" :key="option.value" class="cursor-pointer">
      <input
        type="radio"
        :name="groupName"
        :value="option.value"
        :checked="value === option.value"
        class="peer sr-only"
        @change="emit('update:value', option.value)"
        @keydown.enter.prevent="emit('update:value', option.value)"
      />
      <span
        class="rounded-sm select-none peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-primary"
        :class="
          value === option.value ? 'font-medium text-primary' : 'text-ink'
        "
      >
        {{ option.label }}
      </span>
    </label>
  </div>
</template>
