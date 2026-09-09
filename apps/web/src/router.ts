import { createRouter, createWebHistory } from 'vue-router';

import PostListView from '@/views/posts/PostListView.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: { name: 'posts' } },
    {
      path: '/posts',
      name: 'posts',
      component: PostListView,
      meta: { title: '讨论' },
    },
    { path: '/:pathMatch(.*)*', redirect: { name: 'posts' } },
  ],
});

router.afterEach((to) => {
  document.title = `${String(to.meta.title ?? '社区')} | Novelia Forum`;
});

export default router;
