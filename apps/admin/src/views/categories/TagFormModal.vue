<script setup lang="ts">
import { XButton, XInput } from '@novelia/web-kit';
import { NForm, NFormItem, NInputNumber, NModal } from 'naive-ui';

defineProps<{
  show: boolean;
  editing: boolean;
  saving: boolean;
}>();

const emit = defineEmits<{
  close: [];
  save: [];
}>();

const name = defineModel<string>('name', { required: true });
const color = defineModel<number>('color', { required: true });
const sortOrder = defineModel<number>('sortOrder', { required: true });
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    :title="editing ? '编辑标签' : '新建标签'"
    class="w-11/12! max-w-lg"
    :mask-closable="!saving"
    :close-on-esc="!saving"
    @update:show="(value) => !value && emit('close')"
  >
    <n-form label-placement="top">
      <n-form-item label="标签名称" required>
        <XInput v-model="name" class="w-full" placeholder="输入标签名称" />
      </n-form-item>
      <div class="grid grid-cols-2 gap-3.5">
        <n-form-item label="预定义色号">
          <n-input-number v-model:value="color" :min="0" />
        </n-form-item>
        <n-form-item label="排序值">
          <n-input-number v-model:value="sortOrder" />
        </n-form-item>
      </div>
    </n-form>
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
          :disabled="!name.trim()"
          @click="emit('save')"
        >
          保存
        </XButton>
      </div>
    </template>
  </n-modal>
</template>
