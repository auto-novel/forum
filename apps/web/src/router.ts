import { createRouter, createWebHistory } from 'vue-router';

import PostDetailView from '@/views/post-detail/PostDetailView.vue';
import PostListView from '@/views/post-list/PostListView.vue';

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
    {
      path: '/posts/:id',
      name: 'post-detail',
      component: PostDetailView,
      meta: { title: '帖子详情' },
    },
    { path: '/:pathMatch(.*)*', redirect: { name: 'posts' } },
  ],
});

router.afterEach((to) => {
  document.title = `${String(to.meta.title ?? '社区')} | Novelia Forum`;
});

export default router;
