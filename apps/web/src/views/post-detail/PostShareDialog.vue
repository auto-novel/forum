<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';

import {
  Toast,
  XButton,
  XDialog,
  XInput,
  copyToClipboard,
} from '@novelia/web-kit';
import { postShareMarkdown, postShareUrl } from '@/utils/postShare';

const props = defineProps<{
  open: boolean;
  post: { id: number; title: string };
}>();

const emit = defineEmits<{ 'update:open': [open: boolean] }>();

const linkInput = ref<InstanceType<typeof XInput>>();
const copying = ref(false);

const link = computed(() => postShareUrl(props.post));
const markdown = computed(() => postShareMarkdown(props.post));

// 打开后选中链接，用户也可以直接用系统快捷键复制。
watch(
  () => props.open,
  async (open) => {
    if (!open) return;
    copying.value = false;
    await nextTick();
    selectLink();
  },
);

function focusLink(event: Event) {
  event.preventDefault();
  linkInput.value?.focus();
  selectLink();
}

function selectLink() {
  linkInput.value?.select();
}

async function copy(content: string, successMessage: string) {
  if (copying.value) return;
  copying.value = true;
  await Toast.promise(
    copyToClipboard(content).then(() => emit('update:open', false)),
    {
      success: successMessage,
      error: '复制失败，请手动选中链接复制',
    },
  );
  copying.value = false;
}

function copyLink() {
  void copy(link.value, '帖子链接已复制');
}

function copyMarkdown() {
  void copy(markdown.value, 'Markdown 引用已复制');
}
</script>

<template>
  <XDialog
    :open="open"
    title="分享帖子"
    description="复制链接发送给好友，或复制 Markdown 引用到其他平台。"
    class="w-md"
    @update:open="emit('update:open', $event)"
    @open-auto-focus="focusLink"
  >
    <label class="block text-xs font-medium text-muted" for="post-share-link">
      帖子链接
    </label>
    <XInput
      id="post-share-link"
      ref="linkInput"
      :model-value="link"
      readonly
      class="mt-2"
      @focus="selectLink"
    />

    <div class="mt-5 flex flex-wrap justify-end gap-3">
      <XButton
        color="neutral"
        variant="outline"
        :disabled="copying"
        @click="copyMarkdown"
      >
        复制 Markdown
      </XButton>
      <XButton :disabled="copying" @click="copyLink">复制链接</XButton>
    </div>
  </XDialog>
</template>
