import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { compile, createSSRApp, h } from 'vue';
import { renderToString } from 'vue/server-renderer';

const source = readFileSync(
  new URL('../src/views/post-list/PostListView.vue', import.meta.url),
  'utf8',
);
const template = source.slice(
  source.indexOf('<template>') + 10,
  source.lastIndexOf('</template>'),
);
const render = compile(template);
const passthrough = {
  setup(_, { slots }) {
    return () => h('div', [slots.default?.(), slots.actions?.()]);
  },
};
const pageHeader = {
  props: ['title'],
  setup:
    (props, { slots }) =>
    () =>
      h('header', [h('h1', slots.title?.() ?? props.title), slots.actions?.()]),
};
const dropdownMenu = {
  setup:
    (_, { slots }) =>
    () =>
      h('div', [slots.trigger?.(), slots.default?.()]),
};

for (const mode of ['mobile', 'tablet', 'desktop']) {
  test(`post list uses web-kit viewport for ${mode} navigation and publish label`, async () => {
    assert.match(source, /const \{ isMobile \} = useWebKitViewport\(\)/);
    const app = createSSRApp({
      render,
      setup: () => ({
        isMobile: mode === 'mobile',
        selectedCategory: 'novel',
        selectedCategoryItem: { id: 1, title: '小说' },
        categoryStore: {
          categories: [{ slug: 'novel', title: '小说' }],
          canPublish: () => true,
        },
        searchQuery: '',
        selectedSort: 'updated',
        selectedTagId: undefined,
        applyFilters() {},
        changeCategory() {},
        RouterLink: passthrough,
        posts: [],
        postsLoading: false,
        postsError: undefined,
        page: 1,
        totalPages: 1,
        hasFilters: false,
        loadPosts() {},
        changePage() {},
      }),
    });
    for (const name of ['RouterLink', 'PostFilters', 'XButton', 'XPage']) {
      app.component(name, passthrough);
    }
    app.component('XPageHeader', pageHeader);
    app.component('XDropdownMenu', dropdownMenu);
    app.component('XDropdownMenuItem', passthrough);
    for (const name of [
      'AddOutlined',
      'CheckOutlined',
      'ExpandMoreOutlined',
      'PostList',
    ]) {
      app.component(name, { render: () => null });
    }
    const warnings = [];
    app.config.warnHandler = (message) => warnings.push(message);
    const html = await renderToString(app);
    assert.deepEqual(warnings, []);
    assert.ok(html.includes('<h1>'));
    assert.ok(html.includes('小说'));
    assert.equal(html.includes('aria-label="切换帖子板块"'), mode === 'mobile');
    assert.ok(html.includes('aria-label="发帖"'));
    assert.ok(html.includes('<span>发帖</span>'));
  });
}
