<script setup lang="ts">
import { CloseOutlined } from '@vicons/material';
import { computed, ref, watch } from 'vue';
import {
  DialogContent,
  DialogDescription,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui';

import { authApi } from '@/api';
import AppButton from '@/components/AppButton.vue';
import { notifyError, notifySuccess } from '@/notifications';
import { getApiErrorMessage } from '@/utils/apiError';

const props = defineProps<{
  open: boolean;
  action: 'strike' | 'ban';
  username: string;
  evidence: string;
}>();

const emit = defineEmits<{
  'update:open': [open: boolean];
}>();

const reason = ref('');
const point = ref(1);
const submitting = ref(false);

const isStrike = computed(() => props.action === 'strike');
const title = computed(() =>
  isStrike.value ? `处罚 @${props.username}` : `封禁 @${props.username}`,
);
const description = computed(() =>
  isStrike.value
    ? '处罚会计入该用户的账号记录，请填写原因和处罚分值。'
    : '确认后将封禁该用户的账号，请填写封禁原因。',
);
const canSubmit = computed(
  () =>
    reason.value.trim().length > 0 &&
    (!isStrike.value || (Number.isInteger(point.value) && point.value > 0)),
);

watch(
  () => [props.open, props.action, props.username] as const,
  ([open]) => {
    if (!open) return;
    reason.value = '';
    point.value = 1;
  },
);

function updateOpen(open: boolean) {
  if (!submitting.value) emit('update:open', open);
}

async function submit() {
  const value = reason.value.trim();
  if (!canSubmit.value || submitting.value) return;

  submitting.value = true;
  try {
    if (isStrike.value) {
      await authApi.createStrike({
        username: props.username,
        reason: value,
        evidence: props.evidence,
        point: point.value,
      });
      notifySuccess(`已处罚 @${props.username}`);
    } else {
      await authApi.banUser({ username: props.username, reason: value });
      notifySuccess(`已封禁 @${props.username}`);
    }
    emit('update:open', false);
  } catch (error) {
    notifyError(
      await getApiErrorMessage(
        error,
        isStrike.value ? '处罚用户失败' : '封禁用户失败',
      ),
    );
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <DialogRoot :open="open" @update:open="updateOpen">
    <DialogPortal>
      <DialogOverlay class="fixed inset-0 z-40 bg-black/45" />
      <DialogContent
        class="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-md border border-border bg-surface p-5 shadow-2xl outline-none sm:p-6"
      >
        <DialogTitle class="text-lg font-semibold text-ink">
          {{ title }}
        </DialogTitle>
        <DialogDescription class="mt-2 text-sm leading-6 text-muted">
          {{ description }}
        </DialogDescription>

        <form class="mt-5 space-y-4" @submit.prevent="submit">
          <label class="block">
            <span class="text-sm font-medium text-ink">
              {{ isStrike ? '处罚原因' : '封禁原因' }}
            </span>
            <textarea
              v-model="reason"
              class="mt-1.5 min-h-24 w-full resize-y rounded-sm border border-border bg-surface px-3 py-2 text-sm leading-6 text-ink outline-none transition-colors placeholder:text-muted focus:border-primary focus:ring-1 focus:ring-primary disabled:opacity-50"
              :placeholder="isStrike ? '说明违规行为' : '说明封禁原因'"
              :disabled="submitting"
              required
              autofocus
            />
          </label>

          <label v-if="isStrike" class="block">
            <span class="text-sm font-medium text-ink">处罚分值</span>
            <input
              v-model.number="point"
              type="number"
              min="1"
              step="1"
              class="mt-1.5 block w-28 rounded-sm border border-border bg-surface px-3 py-2 text-sm text-ink outline-none transition-colors focus:border-primary focus:ring-1 focus:ring-primary disabled:opacity-50"
              :disabled="submitting"
              required
            />
          </label>

          <div v-if="isStrike">
            <p class="text-sm font-medium text-ink">处罚依据</p>
            <p
              class="mt-1.5 max-h-28 overflow-y-auto whitespace-pre-wrap break-all rounded-sm bg-paper px-3 py-2 text-xs leading-5 text-muted"
            >
              {{ evidence }}
            </p>
          </div>

          <div class="flex justify-end gap-3 pt-2">
            <AppButton
              variant="outline"
              :disabled="submitting"
              @click="updateOpen(false)"
            >
              取消
            </AppButton>
            <AppButton
              type="submit"
              variant="danger"
              :disabled="!canSubmit || submitting"
            >
              {{ submitting ? '处理中…' : isStrike ? '确认处罚' : '确认封禁' }}
            </AppButton>
          </div>
        </form>

        <AppButton
          class="absolute top-4 right-4"
          variant="subtle"
          size="icon-sm"
          aria-label="关闭"
          :disabled="submitting"
          @click="updateOpen(false)"
        >
          <CloseOutlined class="size-5" aria-hidden="true" />
        </AppButton>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>
