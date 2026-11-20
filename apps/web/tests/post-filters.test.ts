// @vitest-environment jsdom
import { afterEach, expect, test, vi } from 'vite-plus/test';
import { enableAutoUnmount, mount } from '@vue/test-utils';
import PostFilters from '@/views/post-list/PostFilters.vue';

vi.mock('@/stores/category', () => ({
  useCategoryStore: () => ({
    categories: [
      { id: 1, slug: 'novel' },
      { id: 2, slug: 'feedback' },
    ],
    defaultCategory: { slug: 'novel' },
    tagsByCategoryId: (id: number) =>
      id === 1
        ? Array.from({ length: 12 }, (_, i) => ({
            id: i + 1,
            name: `标签${i + 1}`,
          }))
        : [{ id: 20, name: '反馈' }],
  }),
}));
vi.mock('@vicons/material', () => ({ SearchOutlined: { render: () => null } }));

vi.mock('@nuxt/ui/components/Button.vue', () => ({
  default: { template: '<button><slot /></button>' },
}));

enableAutoUnmount(afterEach);
function setup(tagId?: number) {
  return mount(PostFilters, {
    props: { categoryId: 1, query: '', tagId, sort: 'active' },
    global: {
      stubs: {
        UButton: { template: '<button><slot /></button>' },
        UInput: {
          props: ['modelValue'],
          emits: ['update:modelValue'],
          template:
            '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
        },
        USelect: true,
      },
    },
  });
}

test('selected overflow tags survive collapse and route changes sync filters without submitting', async () => {
  const wrapper = setup(12);
  const tag = (name: string) =>
    wrapper.findAll('button').find((button) => button.text() === name)!;
  expect(tag('标签12').attributes('aria-pressed')).toBe('true');
  expect(tag('标签9')).toBeUndefined();
  await tag('更多标签').trigger('click');
  expect(tag('标签9')).toBeDefined();
  await tag('收起').trigger('click');
  expect(tag('标签12').attributes('aria-pressed')).toBe('true');
  await wrapper.setProps({ tagId: 11, query: 'new', sort: 'comments' });
  expect(tag('标签11').attributes('aria-pressed')).toBe('true');
  expect(wrapper.find('input').element.value).toBe('new');
  await tag('更多标签').trigger('click');
  await wrapper.setProps({ categoryId: 2, tagId: undefined });
  expect(tag('全部标签').attributes('aria-pressed')).toBe('true');
  expect(tag('收起')).toBeUndefined();
  expect(wrapper.emitted('apply')).toBeUndefined();
});

test('tag selection applies trimmed filters once, and all clears the tag', async () => {
  const wrapper = setup();
  const tag = (name: string) =>
    wrapper.findAll('button').find((button) => button.text() === name)!;
  await wrapper.find('input').setValue('  keyword  ');
  await tag('标签3').trigger('click');
  await tag('标签3').trigger('click');
  expect(wrapper.emitted('apply')).toEqual([
    [{ category: 'novel', query: 'keyword', tagId: 3, sort: 'active' }],
  ]);
  await tag('全部标签').trigger('click');
  expect(wrapper.emitted('apply')?.[1]).toEqual([
    { category: 'novel', query: 'keyword', tagId: undefined, sort: 'active' },
  ]);
});
