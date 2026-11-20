import { afterEach, beforeEach, expect, test, vi } from 'vite-plus/test';
import type * as VueRouter from 'vue-router';
import router from '@/router';

const mocks = vi.hoisted(() => ({ setPostListReturn: vi.fn() }));
vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof VueRouter>();
  return { ...actual, createWebHistory: actual.createMemoryHistory };
});
vi.mock('@/stores/category', () => ({
  useCategoryStore: () => ({
    getLastVisitedCategory: () => 'novel',
    categories: [{ id: 1, slug: 'novel', title: '小说' }],
    saveLastVisitedCategory() {},
  }),
}));
vi.mock('@/utils/postNavigation', () => ({
  setPostListReturn: mocks.setPostListReturn,
}));
vi.mock('@novelia/web-kit', () => ({
  MyStrikeListView: {},
  restoreScrollPosition: async (position: object, isCurrent: () => boolean) =>
    isCurrent() ? { ...position, behavior: 'instant' } : false,
}));
// Route components are irrelevant to the real router's guards and return-state bookkeeping.
vi.mock('@/views/community-rules/CommunityRulesView.vue', () => ({
  default: {},
}));
vi.mock('@/views/post-list/PersonalPostListView.vue', () => ({ default: {} }));
vi.mock('@/views/post-detail/PostDetailView.vue', () => ({ default: {} }));
vi.mock('@/views/post-editor/PostCreateView.vue', () => ({ default: {} }));
vi.mock('@/views/post-editor/PostEditView.vue', () => ({ default: {} }));
vi.mock('@/views/post-list/PostListView.vue', () => ({ default: {} }));
vi.mock('@/views/profile/ProfileView.vue', () => ({ default: {} }));
vi.mock('@/views/settings/AccountSettingsView.vue', () => ({ default: {} }));
vi.mock('@/views/settings/BlacklistView.vue', () => ({ default: {} }));

beforeEach(async () => {
  vi.stubGlobal('document', { title: '' });
  vi.stubGlobal('window', { scrollY: 120 });
  await router.push('/rules');
  mocks.setPostListReturn.mockClear();
});
afterEach(() => vi.unstubAllGlobals());

test('saved scroll restoration rejects a superseded navigation', async () => {
  const to = router.currentRoute.value;
  const scroll = router.options.scrollBehavior!;
  const position = { top: 120, left: 24 };
  expect(scroll(to, to, null)).toBe(false);
  expect(await scroll(to, to, position)).toEqual({
    ...position,
    behavior: 'instant',
  });
  await router.push('/my/posts');
  expect(await scroll(to, to, position)).toBe(false);
});

test.each([
  ['/my?page=2#saved', '/my/posts?page=2#saved'],
  ['/favorites?page=3#saved', '/my/favorites?page=3#saved'],
  ['/settings?page=2#saved', '/settings/blacklist?page=2#saved'],
])('redirect %s preserves query and hash', async (from, target) => {
  await router.push(from);
  expect(router.currentRoute.value.fullPath).toBe(target);
});

test('list return state survives same-post editing but is cleared by unrelated navigation', async () => {
  for (const path of [
    '/c/novel?page=2',
    '/my/posts?page=2',
    '/my/favorites?page=3',
  ]) {
    await router.push(path);
    await router.push('/p/10');
    expect(mocks.setPostListReturn).toHaveBeenLastCalledWith({
      path,
      scrollTop: 120,
    });
    mocks.setPostListReturn.mockClear();
    await router.push('/p/10/edit');
    await router.push('/p/10');
    expect(mocks.setPostListReturn).not.toHaveBeenCalled();
    await router.push('/p/11');
    expect(mocks.setPostListReturn).toHaveBeenLastCalledWith();
    mocks.setPostListReturn.mockClear();
    await router.push('/rules');
    expect(mocks.setPostListReturn).toHaveBeenLastCalledWith();
  }
});
