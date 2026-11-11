<script setup lang="ts">
import { useId, watch } from 'vue';

const props = defineProps<{
  show: boolean;
  editing: boolean;
  saving: boolean;
}>();

const emit = defineEmits<{
  close: [];
  save: [];
}>();

const fieldId = useId();
let returnFocus: HTMLElement | null = null;
watch(
  () => props.show,
  (open) => {
    if (open)
      returnFocus =
        document.activeElement instanceof HTMLElement
          ? document.activeElement
          : null;
  },
  { flush: 'sync' },
);
const modalContent = {
  onCloseAutoFocus(event: Event) {
    if (!returnFocus?.isConnected) return;
    event.preventDefault();
    returnFocus.focus({ preventScroll: true });
  },
};
const name = defineModel<string>('name', { required: true });
const color = defineModel<number | null>('color', { required: true });
const sortOrder = defineModel<number | null>('sortOrder', { required: true });
</script>

<template>
  <UModal
    :open="show"
    :content="modalContent"
    :title="editing ? '编辑标签' : '新建标签'"
    :ui="{ content: 'w-11/12' }"
    :dismissible="!saving"
    :close="!saving"
    @update:open="(value) => !value && !saving && emit('close')"
  >
    <template #body>
      <div class="grid gap-5">
        <div class="grid gap-2">
          <label :for="`${fieldId}-name`" class="text-sm">
            标签名称
            <span class="text-error" aria-hidden="true">*</span>
          </label>
          <UInput
            :id="`${fieldId}-name`"
            v-model="name"
            class="w-full"
            :disabled="saving"
            placeholder="输入标签名称"
            aria-required="true"
          />
        </div>
        <div class="grid grid-cols-2 gap-3.5">
          <div class="grid gap-2">
            <label :for="`${fieldId}-color`" class="text-sm">预定义色号</label>
            <UInputNumber
              :id="`${fieldId}-color`"
              v-model="color"
              :min="0"
              :disabled="saving"
            />
          </div>
          <div class="grid gap-2">
            <label :for="`${fieldId}-sort-order`" class="text-sm">排序值</label>
            <UInputNumber
              :id="`${fieldId}-sort-order`"
              v-model="sortOrder"
              :disabled="saving"
            />
          </div>
        </div>
      </div>
    </template>
    <template #footer>
      <UButton
        size="lg"
        color="neutral"
        variant="outline"
        :disabled="saving"
        @click="emit('close')"
      >
        取消
      </UButton>
      <UButton
        size="lg"
        :loading="saving"
        :disabled="saving || !name.trim() || color == null || sortOrder == null"
        @click="emit('save')"
      >
        保存
      </UButton>
    </template>
  </UModal>
</template>
