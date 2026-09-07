<script setup lang="ts">
import { NButton, NForm, NFormItem, NInput, NModal } from 'naive-ui';

defineProps<{
  show: boolean;
  editing: boolean;
  saving: boolean;
}>();

const emit = defineEmits<{
  close: [];
  save: [];
}>();

const slug = defineModel<string>('slug', { required: true });
const bannerUrl = defineModel<string>('bannerUrl', { required: true });
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    :title="editing ? '编辑分类' : '新建分类'"
    class="form-modal"
    :mask-closable="!saving"
    :close-on-esc="!saving"
    @update:show="(value) => !value && emit('close')"
  >
    <n-form label-placement="top">
      <n-form-item label="分类标识" required>
        <n-input v-model:value="slug" placeholder="例如 general" />
      </n-form-item>
      <n-form-item label="横幅地址">
        <n-input
          v-model:value="bannerUrl"
          placeholder="https://example.com/banner.webp"
        />
      </n-form-item>
    </n-form>
    <template #footer>
      <div class="modal-actions">
        <n-button :disabled="saving" @click="emit('close')">取消</n-button>
        <n-button
          type="primary"
          :loading="saving"
          :disabled="!slug.trim()"
          @click="emit('save')"
        >
          保存
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.form-modal {
  width: min(520px, calc(100vw - 32px));
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 16px;
}
</style>
