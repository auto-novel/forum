<script setup lang="ts">
import { XButton, XDialog, XInput, XNumberInput } from '@novelia/web-kit';
import { useId } from 'vue';

defineProps<{
  show: boolean;
  editing: boolean;
  saving: boolean;
}>();

const emit = defineEmits<{
  close: [];
  save: [];
}>();

const fieldId = useId();
const name = defineModel<string>('name', { required: true });
const color = defineModel<number | null>('color', { required: true });
const sortOrder = defineModel<number | null>('sortOrder', { required: true });
</script>

<template>
  <XDialog
    :open="show"
    :title="editing ? '编辑标签' : '新建标签'"
    class="w-11/12 max-w-lg"
    :closable="!saving"
    @update:open="(value) => !value && !saving && emit('close')"
  >
    <div class="grid gap-5">
      <div class="grid gap-2">
        <label :for="`${fieldId}-name`" class="text-sm">
          标签名称
          <span class="text-error" aria-hidden="true">*</span>
        </label>
        <XInput
          :id="`${fieldId}-name`"
          v-model="name"
          class="w-full"
          placeholder="输入标签名称"
          aria-required="true"
        />
      </div>
      <div class="grid grid-cols-2 gap-3.5">
        <div class="grid gap-2">
          <label :for="`${fieldId}-color`" class="text-sm">预定义色号</label>
          <XNumberInput :id="`${fieldId}-color`" v-model="color" :min="0" />
        </div>
        <div class="grid gap-2">
          <label :for="`${fieldId}-sort-order`" class="text-sm">排序值</label>
          <XNumberInput :id="`${fieldId}-sort-order`" v-model="sortOrder" />
        </div>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end gap-4">
        <XButton
          color="neutral"
          variant="outline"
          :disabled="saving"
          @click="emit('close')"
        >
          取消
        </XButton>
        <XButton
          :loading="saving"
          :disabled="!name.trim() || color == null || sortOrder == null"
          @click="emit('save')"
        >
          保存
        </XButton>
      </div>
    </template>
  </XDialog>
</template>
