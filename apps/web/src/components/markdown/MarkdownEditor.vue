<script setup lang="ts">
import {
  LinkOutlined,
  StarBorderOutlined,
  UnfoldMoreOutlined,
  VisibilityOffOutlined,
} from '@vicons/material';
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  useTemplateRef,
  watch,
} from 'vue';

import { XButton } from '@novelia/web-kit';

import { handleMarkdownLinkPaste, type MarkdownMode } from '@novelia/forum-api';
import MarkdownContent from './MarkdownContent.vue';

const props = withDefaults(
  defineProps<{
    mode: MarkdownMode;
    placeholder?: string;
    disabled?: boolean;
    describedBy?: string;
    invalid?: boolean;
    maxlength?: number;
    rows?: number;
  }>(),
  {
    placeholder: '使用 Markdown 输入内容…',
    disabled: false,
    maxlength: 100000,
    rows: 7,
  },
);

const value = defineModel<string>({ required: true });
const textarea = useTemplateRef<HTMLTextAreaElement>('textarea');
const activeTab = ref<'edit' | 'preview'>('edit');
const editorTabClass =
  'relative min-w-16 border-r border-border px-[0.9rem] py-[0.55rem] text-[0.8125rem] font-semibold';
// 文章编辑时工具栏吸顶，用哨兵判断是否已经离开滚动区顶部。
const isArticle = computed(() => props.mode === 'article');
const editorRoot = useTemplateRef<HTMLElement>('editorRoot');
const toolbarSentinel = useTemplateRef<HTMLElement>('toolbarSentinel');
const toolbarStuck = ref(false);
let toolbarObserver: IntersectionObserver | undefined;
let widthObserver: ResizeObserver | undefined;
let editorWidth = 0;

// 输入框不出现内部滚动条，也没必要手动缩放：高度始终跟内容走，
// rows 只作为最小高度（height: auto 时的固有高度就是 rows）。
function resizeEditor() {
  const element = textarea.value;
  if (!element || activeTab.value !== 'edit') return;
  // 测量时高度会暂时塌缩，保住窗口位置，避免浏览器夹紧滚动范围后跳动。
  const scrollY = window.scrollY;
  element.style.height = 'auto';
  element.style.height = `${element.scrollHeight}px`;
  if (window.scrollY !== scrollY) {
    window.scrollTo({ top: scrollY, behavior: 'instant' });
  }
}

onMounted(async () => {
  await nextTick();
  resizeEditor();
  observeEditorWidth();
  observeToolbar();
});

onBeforeUnmount(() => {
  toolbarObserver?.disconnect();
  widthObserver?.disconnect();
});

// 换行宽度变化（窗口缩放、侧边栏折叠）会改变实际行数，需要重新测量。
function observeEditorWidth() {
  const element = editorRoot.value;
  if (!element || typeof ResizeObserver === 'undefined') return;
  widthObserver = new ResizeObserver(([entry]) => {
    const width = entry?.contentRect.width ?? 0;
    if (width === editorWidth) return;
    editorWidth = width;
    resizeEditor();
  });
  widthObserver.observe(element);
}

function observeToolbar() {
  if (!isArticle.value) return;
  const sentinel = toolbarSentinel.value;
  if (!sentinel) return;
  toolbarObserver = new IntersectionObserver(
    ([entry]) => {
      // 哨兵在滚动区上方才是真正的吸顶，落在下方（编辑器还在折叠线以下）不算。
      toolbarStuck.value =
        !entry.isIntersecting &&
        entry.boundingClientRect.top < (entry.rootBounds?.top ?? 0);
    },
    // 与工具栏的 sticky 偏移一致，避开 web-kit 的 50px header。
    { rootMargin: '-50px 0px 0px 0px' },
  );
  toolbarObserver.observe(sentinel);
}

watch(value, () => void nextTick(resizeEditor));
watch(activeTab, (tab) => {
  if (tab === 'edit') void nextTick(resizeEditor);
});

async function restoreSelection(start: number, end: number) {
  await nextTick();
  textarea.value?.focus({ preventScroll: true });
  textarea.value?.setSelectionRange(start, end);
}

function wrapSelection(prefix: string, suffix: string, placeholder: string) {
  const element = textarea.value;
  if (!element || props.disabled) return;
  const start = element.selectionStart;
  const end = element.selectionEnd;
  const selected = value.value.slice(start, end) || placeholder;
  value.value = `${value.value.slice(0, start)}${prefix}${selected}${suffix}${value.value.slice(end)}`;
  void restoreSelection(
    start + prefix.length,
    start + prefix.length + selected.length,
  );
}

function insertBlock(prefix: string, suffix: string, placeholder: string) {
  const element = textarea.value;
  if (!element || props.disabled) return;
  const start = element.selectionStart;
  const end = element.selectionEnd;
  const before = value.value.slice(0, start);
  const after = value.value.slice(end);
  const selected = value.value.slice(start, end) || placeholder;
  const leadingBreak = before && !before.endsWith('\n') ? '\n' : '';
  const trailingBreak = after && !after.startsWith('\n') ? '\n' : '';
  const insertion = `${leadingBreak}${prefix}${selected}${suffix}${trailingBreak}`;
  value.value = `${before}${insertion}${after}`;
  const selectionStart = start + leadingBreak.length + prefix.length;
  void restoreSelection(selectionStart, selectionStart + selected.length);
}

function focus() {
  activeTab.value = 'edit';
  void nextTick(() => textarea.value?.focus());
}

defineExpose({ focus });
</script>

<template>
  <div
    ref="editorRoot"
    class="rounded-md border border-border"
    :class="isArticle ? 'relative' : 'overflow-hidden'"
  >
    <span
      v-if="isArticle"
      ref="toolbarSentinel"
      class="pointer-events-none absolute inset-x-0 top-0 h-px"
      aria-hidden="true"
    />
    <div
      class="flex flex-wrap items-stretch border-b border-border"
      :class="
        isArticle
          ? [
              'sticky top-[50px] z-20 rounded-t-md bg-body',
              toolbarStuck && 'editor-toolbar-stuck',
            ]
          : ''
      "
    >
      <div class="flex flex-none" role="tablist" aria-label="Markdown 编辑模式">
        <XButton
          variant="ghost"
          size="sm"
          :class="[
            'min-h-0 rounded-none hover:bg-transparent hover:text-ink',
            editorTabClass,
            activeTab === 'edit' ? '-mb-px text-ink' : 'text-muted',
          ]"
          role="tab"
          :aria-selected="activeTab === 'edit'"
          @click="activeTab = 'edit'"
        >
          编辑
        </XButton>
        <XButton
          variant="ghost"
          size="sm"
          :class="[
            'min-h-0 rounded-none hover:bg-transparent hover:text-ink',
            editorTabClass,
            activeTab === 'preview' ? '-mb-px text-ink' : 'text-muted',
          ]"
          role="tab"
          :aria-selected="activeTab === 'preview'"
          @click="activeTab = 'preview'"
        >
          预览
        </XButton>
      </div>

      <div
        v-if="activeTab === 'edit'"
        class="order-last flex w-full items-center gap-0.5 overflow-x-auto border-t border-border p-1 sm:order-none sm:ml-auto sm:w-auto sm:border-t-0"
        aria-label="Markdown 格式工具"
      >
        <XButton
          variant="ghost"
          size="icon-sm"
          class="font-bold text-ink hover:bg-divider/40 hover:text-ink"
          title="粗体"
          aria-label="粗体"
          @mousedown.prevent="wrapSelection('**', '**', '粗体')"
        >
          B
        </XButton>
        <XButton
          variant="ghost"
          size="icon-sm"
          class="italic text-ink hover:bg-divider/40 hover:text-ink"
          title="斜体"
          aria-label="斜体"
          @mousedown.prevent="wrapSelection('*', '*', '斜体')"
        >
          I
        </XButton>
        <XButton
          variant="ghost"
          size="icon-sm"
          class="line-through text-ink hover:bg-divider/40 hover:text-ink"
          title="删除线"
          aria-label="删除线"
          @mousedown.prevent="wrapSelection('~~', '~~', '删除线')"
        >
          S
        </XButton>
        <XButton
          variant="ghost"
          size="icon-sm"
          class="text-ink hover:bg-divider/40 hover:text-ink"
          title="链接"
          aria-label="链接"
          @mousedown.prevent="wrapSelection('[', '](https://)', '链接文字')"
        >
          <LinkOutlined class="mx-auto size-4" aria-hidden="true" />
        </XButton>
        <XButton
          variant="ghost"
          size="icon-sm"
          class="text-ink hover:bg-divider/40 hover:text-ink"
          title="剧透"
          aria-label="剧透"
          @mousedown.prevent="wrapSelection('!!', '!!', '剧透内容')"
        >
          <VisibilityOffOutlined class="mx-auto size-4" aria-hidden="true" />
        </XButton>
        <XButton
          variant="ghost"
          size="icon-sm"
          class="text-ink hover:bg-divider/40 hover:text-ink"
          title="评分"
          aria-label="评分"
          @mousedown.prevent="insertBlock('::: star 5', '', '')"
        >
          <StarBorderOutlined class="mx-auto size-4" aria-hidden="true" />
        </XButton>
        <XButton
          variant="ghost"
          size="icon-sm"
          class="text-ink hover:bg-divider/40 hover:text-ink"
          title="折叠内容"
          aria-label="折叠内容"
          @mousedown.prevent="
            insertBlock('::: details 点击展开\n', '\n:::', '折叠内容')
          "
        >
          <UnfoldMoreOutlined class="mx-auto size-4" aria-hidden="true" />
        </XButton>
      </div>
    </div>

    <div v-show="activeTab === 'edit'" role="tabpanel">
      <textarea
        ref="textarea"
        v-model="value"
        class="block min-h-24 w-full resize-none overflow-hidden border-0 bg-transparent px-3 py-3 text-sm leading-6 text-ink outline-none placeholder:text-muted/70 disabled:cursor-not-allowed disabled:opacity-50"
        :rows="rows"
        :maxlength="maxlength"
        :placeholder="placeholder"
        :disabled="disabled"
        :aria-describedby="describedBy"
        :aria-invalid="invalid || undefined"
        spellcheck="false"
        @paste="handleMarkdownLinkPaste($event, textarea)"
      />
    </div>

    <div v-if="activeTab === 'preview'" class="min-h-48 p-4" role="tabpanel">
      <MarkdownContent v-if="value.trim()" :mode="mode" :source="value" />
      <p v-else class="text-sm text-muted">没有可预览的内容</p>
    </div>
  </div>
</template>

<style scoped>
/* 吸顶后加一点投影，和正在滚动的正文区分开。 */
.editor-toolbar-stuck {
  box-shadow: 0 4px 10px -6px rgb(0 0 0 / 30%);
}
</style>
