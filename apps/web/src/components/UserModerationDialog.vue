<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useForumApi } from '@/api';
import {
  Toast,
  XAutoComplete,
  XButton,
  XDialog,
  useWebKit,
} from '@novelia/web-kit';
import { useCommentStore } from '@/stores/comment';

const props = defineProps<{
  open: boolean;
  action: 'strike' | 'ban';
  userId: number;
  username: string;
  evidence: string;
  commentId?: number;
}>();

const emit = defineEmits<{
  'update:open': [open: boolean];
  commentsDeleted: [];
  commentDeleted: [];
}>();

const forumApi = useForumApi();
const { createStrike, banUser } = useWebKit();
const reason = ref('');
const commentStore = useCommentStore();
const pointOptions = [1, 2, 3];
const point = ref(1);
const removeComments = ref(false);
const submitting = ref(false);

const isStrike = computed(() => props.action === 'strike');
const reasonPresets = computed(() =>
  isStrike.value ? ['骂人', '灌水', '网赚盘'] : ['营利', '宣扬纳粹'],
);
const title = computed(() =>
  isStrike.value ? `处罚 @${props.username}` : `封禁 @${props.username}`,
);
const canSubmit = computed(
  () =>
    reason.value.trim().length > 0 &&
    (!isStrike.value || pointOptions.includes(point.value)),
);

watch(
  () => [props.open, props.action, props.username] as const,
  ([open]) => {
    if (!open) return;
    reason.value = '';
    point.value = 1;
    removeComments.value = false;
  },
);

function updateOpen(open: boolean) {
  if (!submitting.value) emit('update:open', open);
}

async function submit() {
  const value = reason.value.trim();
  if (!canSubmit.value || submitting.value) return;

  submitting.value = true;
  const strike = isStrike.value;
  const username = props.username;
  const userId = props.userId;
  const commentId = strike ? props.commentId : undefined;
  const shouldRemoveComments = !strike && removeComments.value;
  const request = strike
    ? createStrike({
        username,
        reason: value,
        evidence: props.evidence,
        point: point.value,
      })
    : banUser({ username, reason: value });
  let succeeded = false;
  try {
    await Toast.promise(
      request.then(() => {
        succeeded = true;
      }),
      {
        success: strike ? `已处罚 @${username}` : `已封禁 @${username}`,
        error: strike ? '处罚用户失败' : '封禁用户失败',
      },
    );
    if (!succeeded) return;

    try {
      if (commentId != null) {
        await Toast.promise(
          commentStore.deleteComment(commentId, true).then(() => {
            emit('commentDeleted');
          }),
          {
            success: '已删除被处罚的评论',
            error: '处罚已成功，但删除评论失败，请手动删除',
          },
        );
      } else if (shouldRemoveComments) {
        await Toast.promise(
          forumApi.deleteCommentsByAuthor(userId).then(() => {
            commentStore.registerDeletedCommentsByAuthor(userId);
            emit('commentsDeleted');
          }),
          {
            success: `已删除 @${username} 的全部评论`,
            error: '删除该用户的评论失败',
          },
        );
      }
    } finally {
      emit('update:open', false);
    }
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <XDialog
    :open="open"
    :title="title"
    :closable="!submitting"
    class="w-[28rem]"
    @update:open="updateOpen"
  >
    <form class="space-y-4" @submit.prevent="submit">
      <label class="block">
        <span class="text-sm font-medium text-ink">
          {{ isStrike ? '处罚原因' : '封禁原因' }}
        </span>
        <XAutoComplete
          v-model="reason"
          :options="reasonPresets"
          size="sm"
          class="mt-1.5 bg-surface"
          :placeholder="isStrike ? '选择或填写处罚原因' : '选择或填写封禁原因'"
          empty-text="无匹配原因，可直接填写"
          :disabled="submitting"
          required
          autofocus
        />
      </label>

      <fieldset v-if="isStrike" :disabled="submitting">
        <legend class="text-sm font-medium text-ink">处罚分值</legend>
        <div class="mt-1.5 flex gap-2">
          <label v-for="option in pointOptions" :key="option">
            <input
              v-model="point"
              type="radio"
              name="strike-point"
              :value="option"
              class="peer sr-only"
            />
            <span
              class="block cursor-pointer rounded-sm border border-border bg-surface px-4 py-1.5 text-sm text-muted transition-colors peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-disabled:cursor-not-allowed peer-disabled:opacity-50"
              :class="{
                'hover:border-primary hover:text-primary peer-checked:border-primary peer-checked:bg-primary/10 peer-checked:text-primary peer-focus-visible:outline-primary':
                  option === 1,
                'hover:border-warning hover:text-warning peer-checked:border-warning peer-checked:bg-warning-soft peer-checked:text-warning peer-focus-visible:outline-warning':
                  option === 2,
                'hover:border-error hover:text-error peer-checked:border-error peer-checked:bg-error-soft peer-checked:text-error peer-focus-visible:outline-error':
                  option === 3,
              }"
            >
              {{ option }} 分
            </span>
          </label>
        </div>
      </fieldset>

      <label
        v-else
        class="flex cursor-pointer items-start gap-3 rounded-sm border border-border bg-paper/60 px-3 py-3"
      >
        <input
          v-model="removeComments"
          type="checkbox"
          class="mt-0.5 size-4 shrink-0 accent-primary"
          :disabled="submitting"
        />
        <span>
          <span class="block text-sm font-medium text-ink">
            删除该用户的全部评论
          </span>
          <span class="mt-1 block text-xs leading-5 text-muted">
            包括该用户在所有帖子及外部资源下发表的评论。
          </span>
        </span>
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
        <XButton
          color="neutral"
          variant="outline"
          :disabled="submitting"
          @click="updateOpen(false)"
        >
          取消
        </XButton>
        <XButton
          type="submit"
          color="danger"
          :disabled="!canSubmit || submitting"
        >
          {{ submitting ? '处理中…' : isStrike ? '确认处罚' : '确认封禁' }}
        </XButton>
      </div>
    </form>
  </XDialog>
</template>
