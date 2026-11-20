import { afterEach, expect, test, vi } from 'vite-plus/test';
import { effectScope, type EffectScope } from 'vue';
import { useBlacklistManagement } from '../src/views/settings/useBlacklistManagement';

const fixture = await vi.hoisted(async () => {
  const { ref } = await import('vue');
  return {
    whoami: ref<{ user: { id: number } | undefined; isSignedIn: boolean }>({
      user: undefined,
      isSignedIn: false,
    }),
    getBlacklist: vi.fn<(signal: AbortSignal) => Promise<{ items: Entry[] }>>(),
    setUserBlocked:
      vi.fn<(userId: number, blocked: boolean) => Promise<void>>(),
    success: vi.fn<(message: string) => void>(),
    error: vi.fn<(message: string) => void>(),
    postCache: vi.fn<(userId: number, blocked: boolean) => void>(),
    commentCache: vi.fn<(userId: number, blocked: boolean) => void>(),
  };
});

vi.mock('@/api', () => ({
  useForumApi: () => ({
    getBlacklist: fixture.getBlacklist,
    setUserBlocked: fixture.setUserBlocked,
  }),
}));
vi.mock('@novelia/web-kit', () => ({
  useWebKit: () => ({ whoami: fixture.whoami }),
  getApiErrorMessage: async (_: unknown, fallback: string) => fallback,
  Toast: { success: fixture.success, error: fixture.error },
}));
vi.mock('@/stores/post', () => ({
  usePostStore: () => ({ refreshBlacklist: fixture.postCache }),
}));
vi.mock('@/stores/comment', () => ({
  useCommentStore: () => ({ refreshBlacklist: fixture.commentCache }),
}));

type Entry = { userId: number; createdAt: string };
const entries: Entry[] = [10, 11].map((userId) => ({
  userId,
  createdAt: '2026-01-01T00:00:00Z',
}));
const scopes: EffectScope[] = [];

afterEach(() => {
  for (const scope of scopes.splice(0)) scope.stop();
  vi.resetAllMocks();
});

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}

function nextList() {
  const request = deferred<{ items: Entry[] }>();
  fixture.getBlacklist.mockReturnValueOnce(request.promise);
  return request;
}

function nextRemoval() {
  const request = deferred<void>();
  fixture.setUserBlocked.mockReturnValueOnce(request.promise);
  return request;
}

function signIn(id: number | undefined) {
  fixture.whoami.value = {
    user: id === undefined ? undefined : { id },
    isSignedIn: id !== undefined,
  };
}

function setup(viewerId: number | null = 7) {
  signIn(viewerId ?? undefined);
  const initial = viewerId === null ? undefined : nextList();
  const scope = effectScope();
  scopes.push(scope);
  const management = scope.run(useBlacklistManagement)!;
  return { scope, management, initial };
}

async function flush() {
  // Include the asynchronous error formatter between catch and finally.
  for (let i = 0; i < 5; i++) await Promise.resolve();
}

async function loaded() {
  const context = setup();
  context.initial!.resolve({ items: entries });
  await flush();
  return context;
}

function expectCaches(userId?: number) {
  for (const cache of [fixture.postCache, fixture.commentCache]) {
    if (userId === undefined) expect(cache).not.toHaveBeenCalled();
    else expect(cache.mock.calls).toEqual([[userId, false]]);
  }
}

test('guests cannot load or remove; signing in loads immediately with an abort signal', async () => {
  const { management } = setup(null);
  await management.load();
  management.selected.value = entries[0];
  await management.confirmRemove();
  expect(fixture.getBlacklist).not.toHaveBeenCalled();
  expect(fixture.setUserBlocked).not.toHaveBeenCalled();
  expect(management.loading.value).toBe(false);

  const request = nextList();
  signIn(7);
  expect(fixture.getBlacklist).toHaveBeenCalledTimes(1);
  expect(fixture.getBlacklist.mock.calls[0]![0]).toBeInstanceOf(AbortSignal);
  expect(fixture.getBlacklist.mock.calls[0]![0].aborted).toBe(false);
  expect(management.loading.value).toBe(true);
  expect(management.selected.value).toBeUndefined();
  request.resolve({ items: entries });
  await flush();
  expect(management.entries.value).toEqual(entries);
  expect(management.loading.value).toBe(false);
});

test('load failure exposes an error; retry clears it, aborts the previous request and replaces entries', async () => {
  const { management, initial } = setup();
  expect(management.loading.value).toBe(true);
  initial!.reject(new Error('network'));
  await flush();
  expect(management.error.value).toBe('加载黑名单失败');
  expect(management.loading.value).toBe(false);
  expect(management.entries.value).toEqual([]);
  const retry = nextList();
  const pending = management.load();
  expect(management.error.value).toBe('');
  expect(management.loading.value).toBe(true);
  expect(fixture.getBlacklist.mock.calls[0]![0].aborted).toBe(true);
  retry.resolve({ items: entries });
  await pending;
  expect(management.entries.value).toEqual(entries);
  expect(management.loading.value).toBe(false);
  expect(fixture.error).not.toHaveBeenCalled();
});

test('removal locks duplicate submissions, refreshes both caches and removes only the selected entry', async () => {
  const { management } = await loaded();
  await management.confirmRemove();
  expect(fixture.setUserBlocked).not.toHaveBeenCalled();
  management.selected.value = entries[0];
  const request = nextRemoval();
  const pending = management.confirmRemove();
  await management.confirmRemove();
  expect(fixture.setUserBlocked.mock.calls).toEqual([[10, false]]);
  expect(management.removing.value).toBe(true);
  expect(management.entries.value).toEqual(entries);
  request.resolve();
  await pending;
  expectCaches(10);
  expect(management.entries.value).toEqual([entries[1]]);
  expect(management.selected.value).toBeUndefined();
  expect(management.removing.value).toBe(false);
  expect(fixture.success.mock.calls).toEqual([['已取消拉黑']]);
});

test('failed removal preserves data and caches, releases the lock and permits retry', async () => {
  const { management } = await loaded();
  management.selected.value = entries[0];
  const request = nextRemoval();
  const pending = management.confirmRemove();
  request.reject(new Error('network'));
  await pending;
  expect(management.entries.value).toEqual(entries);
  expect(management.selected.value).toEqual(entries[0]);
  expect(management.removing.value).toBe(false);
  expectCaches();
  expect(fixture.error.mock.calls).toEqual([['取消拉黑失败']]);
  const retry = nextRemoval();
  const retried = management.confirmRemove();
  retry.resolve();
  await retried;
  expect(fixture.setUserBlocked).toHaveBeenCalledTimes(2);
  expect(management.entries.value).toEqual([entries[1]]);
  expectCaches(10);
});

test('successful removal invalidates an older pending list so it cannot restore the removed entry', async () => {
  const { management } = await loaded();
  const list = nextList();
  const loading = management.load();
  const signal = fixture.getBlacklist.mock.calls.at(-1)![0];
  management.selected.value = entries[0];
  const request = nextRemoval();
  const pending = management.confirmRemove();
  request.resolve();
  await pending;
  expect(signal.aborted).toBe(true);
  expect(management.loading.value).toBe(false);
  list.resolve({ items: entries });
  await loading;
  expect(management.entries.value).toEqual([entries[1]]);
});

test.each([false, true])(
  'account switch ignores stale list completion (failure=%s) without clearing the new loading state',
  async (fails) => {
    const { management, initial } = setup();
    management.selected.value = entries[0];
    const current = nextList();
    signIn(8);
    expect(fixture.getBlacklist.mock.calls[0]![0].aborted).toBe(true);
    expect(management.entries.value).toEqual([]);
    expect(management.selected.value).toBeUndefined();
    if (fails) initial!.reject(new Error('old network error'));
    else initial!.resolve({ items: entries });
    await flush();
    expect(management.error.value).toBe('');
    expect(management.entries.value).toEqual([]);
    expect(management.loading.value).toBe(true);
    current.resolve({ items: [entries[1]!] });
    await flush();
    expect(management.entries.value).toEqual([entries[1]]);
    expect(management.loading.value).toBe(false);
  },
);

test.each([false, true])(
  'switching away and back ignores stale removal (failure=%s), list and notifications while preserving the current lock',
  async (fails) => {
    const { management } = await loaded();
    management.selected.value = entries[0];
    const oldRemove = nextRemoval();
    const oldPending = management.confirmRemove();
    const awayList = nextList();
    signIn(8);
    expect(management.removing.value).toBe(false);
    const returnList = nextList();
    signIn(7);
    awayList.resolve({ items: [entries[1]!] });
    await flush();
    expect(management.entries.value).toEqual([]);
    expect(management.loading.value).toBe(true);
    returnList.resolve({ items: entries });
    await flush();
    management.selected.value = entries[1];
    const currentRemove = nextRemoval();
    const currentPending = management.confirmRemove();
    if (fails) oldRemove.reject(new Error('old removal error'));
    else oldRemove.resolve();
    await oldPending;
    expect(management.entries.value).toEqual(entries);
    expect(management.selected.value).toEqual(entries[1]);
    expect(management.removing.value).toBe(true);
    expectCaches();
    expect(fixture.success).not.toHaveBeenCalled();
    expect(fixture.error).not.toHaveBeenCalled();
    currentRemove.resolve();
    await currentPending;
    expect(management.entries.value).toEqual([entries[0]]);
    expectCaches(11);
    expect(management.removing.value).toBe(false);
  },
);

test('signing out clears selection and data and ignores a late list response', async () => {
  const { management } = await loaded();
  management.selected.value = entries[0];
  const request = nextList();
  const pending = management.load();
  signIn(undefined);
  expect(management.entries.value).toEqual([]);
  expect(management.selected.value).toBeUndefined();
  expect(management.loading.value).toBe(false);
  expect(fixture.getBlacklist.mock.calls.at(-1)![0].aborted).toBe(true);
  request.resolve({ items: entries });
  await pending;
  expect(management.entries.value).toEqual([]);
});

test.each([false, true])(
  'scope disposal aborts loading, suppresses late completion (failure=%s) and prevents reload',
  async (fails) => {
    const { scope, management, initial } = setup();
    scope.stop();
    expect(fixture.getBlacklist.mock.calls[0]![0].aborted).toBe(true);
    if (fails) initial!.reject(new Error('late load error'));
    else initial!.resolve({ items: entries });
    await flush();
    expect(management.entries.value).toEqual([]);
    expect(management.error.value).toBe('');
    expect(fixture.error).not.toHaveBeenCalled();
    await management.load();
    expect(fixture.getBlacklist).toHaveBeenCalledTimes(1);
  },
);

test.each([false, true])(
  'scope disposal suppresses removal UI effects (failure=%s), but successful updates still refresh both caches',
  async (fails) => {
    const { scope, management } = await loaded();
    management.selected.value = entries[0];
    const request = nextRemoval();
    const pending = management.confirmRemove();
    scope.stop();
    if (fails) request.reject(new Error('late removal error'));
    else request.resolve();
    await pending;
    expect(fixture.success).not.toHaveBeenCalled();
    expect(fixture.error).not.toHaveBeenCalled();
    expect(management.entries.value).toEqual(entries);
    expectCaches(fails ? undefined : 10);
    await management.confirmRemove();
    expect(fixture.setUserBlocked).toHaveBeenCalledTimes(1);
  },
);
