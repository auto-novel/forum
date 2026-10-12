import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(
  new URL('../src/views/post-list/PostList.vue', import.meta.url),
  'utf8',
);
const restoreSource = source.slice(
  source.indexOf('async function restoreScroll()'),
  source.indexOf('onMounted(restoreScroll)'),
);

function fixture({ loading = false, error, scrollTop = 360, nextTick } = {}) {
  const calls = [];
  const props = { loading, error };
  const route = { fullPath: '/c/novel?page=2' };
  const history = {
    state: { postListScrollTop: scrollTop, position: 4, custom: 'keep' },
    replaceState(state, url) {
      this.state = state;
      calls.push(['replace', url]);
    },
  };
  const window = {
    history,
    scrollTo: (options) => calls.push(['scroll', options]),
  };
  const restore = new Function(
    'props',
    'route',
    'window',
    'nextTick',
    `${restoreSource}\nreturn restoreScroll;`,
  )(props, route, window, nextTick ?? (() => Promise.resolve()));
  return { calls, props, route, history, restore };
}

test('returning to a loaded list restores window scroll and consumes only its marker', async () => {
  const f = fixture();
  await f.restore();
  assert.deepEqual(f.calls, [
    ['scroll', { top: 360, behavior: 'instant' }],
    ['replace', ''],
  ]);
  assert.deepEqual(f.history.state, { position: 4, custom: 'keep' });
  await f.restore();
  assert.equal(f.calls.length, 2);
});

test('list scroll restoration waits for successful loading', async () => {
  for (const initial of [{ loading: true }, { error: 'failed' }]) {
    const f = fixture(initial);
    await f.restore();
    assert.deepEqual(f.calls, []);
    assert.equal(f.history.state.postListScrollTop, 360);
    f.props.loading = false;
    f.props.error = undefined;
    await f.restore();
    assert.equal(f.calls[0][0], 'scroll');
  }
});

test('stale list restoration does not scroll a different route', async () => {
  let f;
  f = fixture({
    nextTick: async () => {
      f.route.fullPath = '/p/1';
    },
  });
  await f.restore();
  assert.deepEqual(f.calls, []);
});

test('invalid history offsets are ignored and zero is restored', async () => {
  for (const scrollTop of [undefined, '120', NaN, Infinity]) {
    const f = fixture({ scrollTop });
    if (scrollTop === undefined) delete f.history.state.postListScrollTop;
    await f.restore();
    assert.deepEqual(f.calls, []);
  }
  const f = fixture({ scrollTop: 0 });
  await f.restore();
  assert.deepEqual(f.calls[0], ['scroll', { top: 0, behavior: 'instant' }]);
});
