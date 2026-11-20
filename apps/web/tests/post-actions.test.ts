// @vitest-environment jsdom
import { afterEach, beforeEach, expect, test, vi } from 'vite-plus/test';
import { flushPromises, mount } from '@vue/test-utils';
import { defineComponent, effectScope, h, nextTick, ref, type Ref } from 'vue';
import type { Post } from '@/api';
import PostActions from '@/views/post-detail/PostActions.vue';
import { usePostFavorite } from '@/views/post-detail/usePostFavorite';

type Viewer = { id: number; role: 'admin' | 'member'; adminMode: boolean };
type Session = {
  user: Viewer | undefined;
  isSignedIn: boolean;
  asAdmin: boolean;
};
type MenuItem = { label: string; disabled?: boolean; onSelect: () => unknown };
const mocks = vi.hoisted(() => ({
  api: {
    setPostFavorite: vi.fn(),
    deletePost: vi.fn(),
    lockPost: vi.fn(),
    unlockPost: vi.fn(),
    pinPost: vi.fn(),
    unpinPost: vi.fn(),
    setPostStatus: vi.fn(),
    setUserBlocked: vi.fn(),
    deleteCommentsByAuthor: vi.fn(),
  },
  session: undefined as unknown as Ref<Session>,
  setPost: vi.fn(),
  refreshPosts: vi.fn(),
  refreshComments: vi.fn(),
  registerDeleted: vi.fn(),
  success: vi.fn(),
  error: vi.fn(),
  copy: vi.fn(),
}));
vi.mock('@/api', () => ({ useForumApi: () => mocks.api }));
vi.mock('@/stores/post', () => ({
  usePostStore: () => ({
    setPost: mocks.setPost,
    refreshBlacklist: mocks.refreshPosts,
  }),
}));
vi.mock('@/stores/comment', () => ({
  useCommentStore: () => ({
    refreshBlacklist: mocks.refreshComments,
    registerDeletedCommentsByAuthor: mocks.registerDeleted,
  }),
}));
vi.mock('@novelia/web-kit', () => ({
  useWebKit: () => ({ whoami: mocks.session }),
  copyToClipboard: mocks.copy,
  getApiErrorMessage: (_reason: unknown, fallback: string) => fallback,
  Toast: {
    success: mocks.success,
    error: mocks.error,
    async promise(
      promise: Promise<unknown>,
      messages: {
        success: string;
        error: (reason: unknown) => string;
      },
    ) {
      try {
        await promise;
        mocks.success(messages.success);
      } catch (reason) {
        mocks.error(await messages.error(reason));
        throw reason;
      }
    },
  },
  XConfirmDialog: { name: 'XConfirmDialog', render: () => null },
  UserBlacklistDialog: { name: 'UserBlacklistDialog', render: () => null },
  UserModerationDialog: { name: 'UserModerationDialog', render: () => null },
}));
vi.mock('@vicons/material', () => ({
  ChatBubbleOutlineOutlined: { render: () => null },
  EditOutlined: { render: () => null },
  MoreVertOutlined: { render: () => null },
  ShareOutlined: { render: () => null },
}));
vi.mock('@nuxt/ui/components/Button.vue', () => ({
  default: {
    name: 'UButton',
    template: '<button><slot /></button>',
  },
}));
vi.mock('@nuxt/ui/components/DropdownMenu.vue', () => ({
  default: {
    name: 'UDropdownMenu',
    props: ['items'],
    template: '<div><slot /></div>',
  },
}));
vi.mock('@/views/post-detail/PostFavoriteButton.vue', () => ({
  default: {
    name: 'PostFavoriteButton',
    render: () => null,
  },
}));
vi.mock('@/views/post-detail/PostShareDialog.vue', () => ({
  default: {
    name: 'PostShareDialog',
    render: () => null,
  },
}));

const Menu = defineComponent({
  name: 'MenuStub',
  props: ['items'],
  setup:
    (_, { slots }) =>
    () =>
      h('div', slots.default?.()),
});
const cleanups: (() => void)[] = [];
function deferred() {
  let resolve!: () => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<void>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}
function viewer(user: Viewer | undefined) {
  mocks.session.value = {
    user,
    isSignedIn: !!user,
    asAdmin: !!user && user.role === 'admin' && user.adminMode,
  };
}
function setup(role: Viewer['role'] | null = 'admin') {
  viewer(role ? { id: 7, role, adminMode: role === 'admin' } : undefined);
  const post = ref<Post>({
    id: 10,
    categoryId: 1,
    authorId: 7,
    authorUsername: 'author',
    authorBlocked: false,
    title: 'Post',
    content: '# 原文\n\n**Markdown**',
    createdAt: new Date().toISOString(),
    updatedAt: '',
    activeAt: '',
    favorited: false,
    commentsLocked: false,
    pinOrder: null,
    commentsCount: 3,
    viewsCount: 0,
    status: 0,
    tags: [],
  });
  mocks.setPost.mockImplementation((value: Post) => {
    post.value = value;
  });
  const scope = effectScope();
  const favorite = scope.run(() => usePostFavorite(post))!;
  const Host = defineComponent({
    setup: () => () =>
      h(PostActions, {
        post: post.value,
        canFavorite: favorite.canFavorite.value,
        favoriteLoading: favorite.loading.value,
        onUpdated: (patch) => {
          post.value = { ...post.value, ...patch };
        },
        onFavorite: favorite.toggle,
      }),
  });
  const wrapper = mount(Host, {
    global: {
      stubs: {
        UDropdownMenu: Menu,
        UButton: { template: '<button><slot /></button>' },
        PostShareDialog: true,
        PostFavoriteButton: true,
        XConfirmDialog: true,
        UserBlacklistDialog: true,
        UserModerationDialog: true,
      },
    },
  });
  const actions = wrapper.getComponent(PostActions);
  const items = () =>
    wrapper.getComponent(Menu).props('items').flat() as MenuItem[];
  const item = (label: string) => {
    const found = items().find((entry) => entry.label === label);
    expect(found, `menu contains ${label}`).toBeDefined();
    return found!;
  };
  const select = (label: string) => item(label).onSelect();
  let disposed = false;
  const dispose = () => {
    if (disposed) return;
    disposed = true;
    wrapper.unmount();
    scope.stop();
  };
  cleanups.push(dispose);
  return { post, favorite, actions, wrapper, item, items, select, dispose };
}
beforeEach(() => {
  vi.resetAllMocks();
  mocks.session = ref<Session>({
    user: undefined,
    isSignedIn: false,
    asAdmin: false,
  });
});
afterEach(() => {
  cleanups.splice(0).forEach((cleanup) => cleanup());
});

test('guest can copy raw Markdown; duplicate copies are locked and failures allow retry', async () => {
  const f = setup(null);
  const request = deferred();
  mocks.copy.mockReturnValueOnce(request.promise).mockResolvedValue(undefined);
  const pending = f.select('复制原文');
  await nextTick();
  expect(f.item('复制原文').disabled).toBe(true);
  await f.select('复制原文');
  expect(mocks.copy).toHaveBeenCalledExactlyOnceWith(f.post.value.content);
  request.reject(new Error('clipboard denied'));
  await pending;
  await nextTick();
  expect(mocks.error).toHaveBeenCalledWith('复制失败，请手动选择帖子内容复制');
  expect(f.item('复制原文').disabled).toBe(false);
  await f.select('复制原文');
  expect(mocks.success).toHaveBeenCalledWith('帖子原文已复制');
});

test('favorite requests share one lock, preserve data on failure and allow retry', async () => {
  const f = setup();
  const failed = deferred();
  mocks.api.setPostFavorite
    .mockReturnValueOnce(failed.promise)
    .mockResolvedValue(undefined);
  const pending = f.favorite.toggle();
  await f.favorite.toggle();
  expect(mocks.api.setPostFavorite).toHaveBeenCalledExactlyOnceWith(10, true);
  expect(f.favorite.loading.value).toBe(true);
  failed.reject(new Error('network'));
  await pending;
  expect(f.post.value.favorited).toBe(false);
  expect(f.favorite.loading.value).toBe(false);
  expect(mocks.error).toHaveBeenCalledWith('更新收藏失败');
  await f.favorite.toggle();
  expect(f.post.value.favorited).toBe(true);
});

test('guest cannot favorite', async () => {
  const f = setup(null);
  await f.favorite.toggle();
  expect(f.favorite.canFavorite.value).toBe(false);
  expect(mocks.api.setPostFavorite).not.toHaveBeenCalled();
});

test('returning to the same post ignores old favorite response without releasing the new lock', async () => {
  const f = setup();
  const old = deferred(),
    current = deferred();
  mocks.api.setPostFavorite
    .mockReturnValueOnce(old.promise)
    .mockReturnValueOnce(current.promise);
  const first = f.favorite.toggle();
  const original = f.post.value;
  f.post.value = { ...original, id: 11 };
  f.post.value = original;
  const second = f.favorite.toggle();
  old.resolve();
  await first;
  expect(f.post.value.favorited).toBe(false);
  expect(f.favorite.loading.value).toBe(true);
  expect(mocks.success).not.toHaveBeenCalled();
  current.resolve();
  await second;
  expect(f.post.value.favorited).toBe(true);
});

test.each([
  ['viewer', false],
  ['viewer', true],
  ['dispose', false],
  ['dispose', true],
] as const)(
  '%s invalidates pending favorite and management results (failure: %s)',
  async (change, fails) => {
    const f = setup();
    const favorite = deferred(),
      management = deferred();
    mocks.api.setPostFavorite.mockReturnValue(favorite.promise);
    mocks.api.pinPost.mockReturnValue(management.promise);
    const pending = f.favorite.toggle();
    f.select('置顶帖子');
    if (change === 'viewer') viewer({ id: 8, role: 'admin', adminMode: true });
    else f.dispose();
    if (fails) {
      favorite.reject(new Error('late failure'));
      management.reject(new Error('late management failure'));
    } else {
      favorite.resolve();
      management.resolve();
    }
    await pending;
    await flushPromises();
    expect(f.post.value.favorited).toBe(false);
    expect(f.actions.emitted('updated')).toBeUndefined();
    expect(mocks.success).not.toHaveBeenCalled();
    expect(mocks.error).not.toHaveBeenCalled();
  },
);

test('admin-mode toggle does not invalidate or release the favorite lock', async () => {
  const f = setup();
  const request = deferred();
  mocks.api.setPostFavorite.mockReturnValue(request.promise);
  const pending = f.favorite.toggle();
  viewer({ id: 7, role: 'admin', adminMode: false });
  await f.favorite.toggle();
  expect(f.favorite.loading.value).toBe(true);
  expect(mocks.api.setPostFavorite).toHaveBeenCalledTimes(1);
  request.resolve();
  await pending;
  expect(f.post.value.favorited).toBe(true);
});

test('management acquires one shared lock before starting requests and retries after failure', async () => {
  const f = setup();
  const request = deferred();
  mocks.api.pinPost
    .mockReturnValueOnce(request.promise)
    .mockResolvedValue(undefined);
  f.select('置顶帖子');
  // Call the public menu handler again even before Vue has rendered disabled state.
  f.select('锁定评论');
  expect(mocks.api.lockPost).not.toHaveBeenCalled();
  await nextTick();
  expect(f.item('锁定评论').disabled).toBe(true);
  request.reject(new Error('network'));
  await flushPromises();
  expect(f.actions.emitted('updated')).toBeUndefined();
  expect(f.item('置顶帖子').disabled).toBe(false);
  expect(mocks.error).toHaveBeenCalledWith('置顶帖子失败');
  f.select('置顶帖子');
  await flushPromises();
  expect(f.actions.emitted('updated')).toEqual([[{ pinOrder: 0 }]]);
});

test.each([true, false])(
  'favorite and management preserve each other and fresh comments (favorite first: %s)',
  async (favoriteFirst) => {
    const f = setup();
    const favorite = deferred(),
      management = deferred();
    mocks.api.setPostFavorite.mockReturnValue(favorite.promise);
    mocks.api.lockPost.mockReturnValue(management.promise);
    const pending = f.favorite.toggle();
    f.select('锁定评论');
    f.post.value = { ...f.post.value, commentsCount: 4 };
    (favoriteFirst ? favorite : management).resolve();
    await flushPromises();
    (favoriteFirst ? management : favorite).resolve();
    await pending;
    await flushPromises();
    expect(f.post.value).toMatchObject({
      favorited: true,
      commentsLocked: true,
      commentsCount: 4,
    });
  },
);

test('management returning to the same post ignores stale responses and preserves current lock', async () => {
  const f = setup();
  const old = deferred(),
    current = deferred();
  mocks.api.pinPost
    .mockReturnValueOnce(old.promise)
    .mockReturnValueOnce(current.promise);
  f.select('置顶帖子');
  const original = f.post.value;
  f.post.value = { ...original, id: 11 };
  await nextTick();
  f.post.value = original;
  await nextTick();
  f.select('置顶帖子');
  old.resolve();
  await flushPromises();
  expect(f.actions.emitted('updated')).toBeUndefined();
  expect(f.item('置顶帖子').disabled).toBe(true);
  current.resolve();
  await flushPromises();
  expect(f.actions.emitted('updated')).toEqual([[{ pinOrder: 0 }]]);
});

test.each([
  ['删除帖子', 'deletePost', [10]],
  ['隐藏帖子', 'setPostStatus', [10, 1]],
] as const)(
  '%s requires confirmation and emits deleted only on success',
  async (label, method, args) => {
    const f = setup();
    const request = deferred();
    mocks.api[method].mockReturnValue(request.promise);
    f.select(label);
    await nextTick();
    expect(mocks.api[method]).not.toHaveBeenCalled();
    const dialog = f.wrapper.getComponent({ name: 'XConfirmDialog' });
    expect(dialog.attributes('open')).toBe('true');
    dialog.vm.$emit('confirm');
    expect(mocks.api[method]).toHaveBeenCalledWith(...args);
    expect(f.actions.emitted('deleted')).toBeUndefined();
    request.resolve();
    await flushPromises();
    expect(f.actions.emitted('deleted')).toEqual([[]]);
  },
);

test('permissions distinguish recent owner, expired owner and admin-mode enabled moderator', async () => {
  const f = setup('member');
  expect(f.wrapper.text()).toContain('编辑');
  expect(f.items().map((entry) => entry.label)).toContain('删除帖子');
  f.post.value = { ...f.post.value, createdAt: '2000-01-01T00:00:00Z' };
  await nextTick();
  expect(f.wrapper.text()).toContain('编辑');
  expect(f.items().map((entry) => entry.label)).not.toContain('删除帖子');
  viewer({ id: 8, role: 'admin', adminMode: false });
  await nextTick();
  expect(f.wrapper.text()).not.toContain('编辑');
  expect(f.items().map((entry) => entry.label)).toEqual([
    '复制原文',
    '屏蔽用户',
  ]);
  viewer({ id: 8, role: 'admin', adminMode: true });
  await nextTick();
  expect(f.wrapper.text()).toContain('编辑');
  expect(f.items().map((entry) => entry.label)).toEqual(
    expect.arrayContaining([
      '置顶帖子',
      '锁定评论',
      '隐藏帖子',
      '删除帖子',
      '处罚用户',
      '封禁用户',
    ]),
  );
});
