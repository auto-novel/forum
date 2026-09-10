import { createRouter, createWebHistory } from 'vue-router';

import FavoritePostListView from '@/views/post-list/FavoritePostListView.vue';
import MyPostListView from '@/views/post-list/MyPostListView.vue';
import PostDetailView from '@/views/post-detail/PostDetailView.vue';
import PostCreateView from '@/views/post-create/PostCreateView.vue';
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
      path: '/favorites',
      name: 'favorites',
      component: FavoritePostListView,
      meta: { title: '我的收藏' },
    },
    {
      path: '/my/posts',
      name: 'my-posts',
      component: MyPostListView,
      meta: { title: '我的帖子' },
    },
    {
      path: '/posts/new',
      name: 'post-create',
      component: PostCreateView,
      meta: { title: '发表帖子' },
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
