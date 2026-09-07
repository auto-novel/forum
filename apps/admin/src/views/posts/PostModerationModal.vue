<script setup lang="ts">
import {
  NButton,
  NForm,
  NFormItem,
  NInputNumber,
  NModal,
  NSelect,
  NSwitch,
  NText,
} from 'naive-ui';
import { reactive, watch } from 'vue';

import { useForumApi, type Post } from '@/api';

const props = defineProps<{ post: Post | null }>();
const emit = defineEmits<{
  close: [];
  success: [message: string];
  error: [message: string];
}>();

const api = useForumApi();
const saving = defineModel<boolean>('saving', { default: false });
const form = reactive({
  status: 0,
  commentsLocked: false,
  pinOrder: null as number | null,
});

watch(
  () => props.post,
  (post) => {
    if (!post) return;
    Object.assign(form, {
      status: post.status,
      commentsLocked: post.commentsLocked,
      pinOrder: post.pinOrder ?? null,
    });
  },
  { immediate: true },
);

function close() {
  if (!saving.value) emit('close');
}

async function save() {
  if (!props.post) return;
  saving.value = true;
  try {
    await api.moderatePost(props.post.id, form);
    emit('success', `帖子「${props.post.title}」已更新`);
  } catch (error) {
    emit('error', error instanceof Error ? error.message : String(error));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <n-modal
    :show="post !== null"
    preset="card"
    title="管理帖子"
    class="form-modal"
    :mask-closable="!saving"
    :close-on-esc="!saving"
    @update:show="(show) => !show && close()"
  >
    <n-text v-if="post" strong class="post-title">{{ post.title }}</n-text>
    <n-form label-placement="top">
      <n-form-item label="帖子状态">
        <n-select
          v-model:value="form.status"
          :options="[
            { label: '正常发布', value: 0 },
            { label: '隐藏', value: 1 },
            { label: '删除', value: 2 },
          ]"
        />
      </n-form-item>
      <n-form-item label="置顶顺序">
        <n-input-number
          v-model:value="form.pinOrder"
          clearable
          placeholder="留空表示不置顶"
        />
      </n-form-item>
      <n-form-item label="锁定评论区">
        <n-switch v-model:value="form.commentsLocked" />
      </n-form-item>
    </n-form>
    <template #footer>
      <div class="modal-actions">
        <n-button :disabled="saving" @click="close">取消</n-button>
        <n-button type="primary" :loading="saving" @click="save">
          保存设置
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.form-modal {
  width: min(520px, calc(100vw - 32px));
}

.post-title {
  display: block;
  margin-bottom: 18px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
