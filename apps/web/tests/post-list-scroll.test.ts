// @vitest-environment jsdom
import { afterEach, beforeEach, expect, test, vi } from 'vite-plus/test';
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils';
import PostList from '@/views/post-list/PostList.vue';

const mocks = await vi.hoisted(async () => {
  const { reactive } = await import('vue');
  return { route: reactive({ fullPath: '/c/novel?page=2' }), restore: vi.fn() };
});
vi.mock('vue-router', () => ({ useRoute: () => mocks.route }));
vi.mock('@novelia/web-kit', () => ({
  restoreScrollPosition: mocks.restore,
  XPagination: { render: () => null },
}));
vi.mock('@/views/post-list/PostListItem.vue', () => ({
  default: { render: () => null },
}));
enableAutoUnmount(afterEach);
beforeEach(() => {
  mocks.route.fullPath = '/c/novel?page=2';
  mocks.restore.mockReset().mockImplementation(async (position, isCurrent) => {
    await Promise.resolve();
    return isCurrent() ? { ...position, behavior: 'instant' } : false;
  });
  window.history.replaceState(
    { postListScrollTop: 360, position: 4, custom: 'keep' },
    '',
  );
  vi.spyOn(window, 'scrollTo').mockImplementation(() => {});
});
function setup(loading = false, error?: string) {
  return mount(PostList, {
    props: { posts: [], loading, error, page: 2, totalPages: 3 },
    global: {
      stubs: { USeparator: true, USkeleton: true, UError: true, UEmpty: true },
    },
  });
}

test('restoration waits for successful loading and consumes only its own history marker', async () => {
  const wrapper = setup(true);
  await flushPromises();
  expect(mocks.restore).not.toHaveBeenCalled();
  await wrapper.setProps({ loading: false, error: 'failed' });
  await flushPromises();
  expect(mocks.restore).not.toHaveBeenCalled();
  await wrapper.setProps({ loading: true, error: undefined });
  await wrapper.setProps({ loading: false });
  await flushPromises();
  expect(window.scrollTo).toHaveBeenCalledExactlyOnceWith({
    top: 360,
    left: 0,
    behavior: 'instant',
  });
  expect(window.history.state).toEqual({ position: 4, custom: 'keep' });
  await wrapper.setProps({ loading: true });
  await wrapper.setProps({ loading: false });
  await flushPromises();
  expect(window.scrollTo).toHaveBeenCalledTimes(1);
});

test.each(['route', 'history', 'marker', 'loading', 'error'] as const)(
  '%s changes cancel in-flight restoration without consuming the current marker',
  async (change) => {
    let isCurrent!: () => boolean;
    let resolve!: (position: false | { top: number; left: number }) => void;
    mocks.restore.mockImplementation((_position, valid) => {
      isCurrent = valid;
      return new Promise((done) => {
        resolve = done;
      });
    });
    const wrapper = setup();
    if (change === 'route') mocks.route.fullPath = '/p/1';
    if (change === 'history')
      window.history.replaceState({ ...window.history.state, position: 5 }, '');
    if (change === 'marker')
      window.history.replaceState(
        { ...window.history.state, postListScrollTop: 99 },
        '',
      );
    if (change === 'loading') await wrapper.setProps({ loading: true });
    if (change === 'error') await wrapper.setProps({ error: 'failed' });
    expect(isCurrent()).toBe(false);
    resolve(isCurrent() ? { top: 360, left: 0 } : false);
    await flushPromises();
    expect(window.scrollTo).not.toHaveBeenCalled();
    expect(window.history.state.postListScrollTop).toBe(
      change === 'marker' ? 99 : 360,
    );
  },
);

test('invalid history offsets are ignored but zero is restored', async () => {
  for (const offset of [undefined, '120', NaN, Infinity, 0]) {
    window.history.replaceState({ postListScrollTop: offset }, '');
    const wrapper = setup();
    await flushPromises();
    wrapper.unmount();
  }
  expect(mocks.restore).toHaveBeenCalledTimes(1);
  expect(window.scrollTo).toHaveBeenCalledExactlyOnceWith({
    top: 0,
    left: 0,
    behavior: 'instant',
  });
});
