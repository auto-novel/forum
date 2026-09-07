import { createRouter, createWebHistory } from 'vue-router';

import { AdminLoginView } from '@novelia/admin-kit';

const APP_TITLE = '论坛服务管理后台';

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: AdminLoginView,
      meta: { title: '登录', guestOnly: true },
    },
    {
      path: '/',
      redirect: { name: 'overview' },
      meta: { requiresAuth: true },
      children: [
        {
          path: 'overview',
          name: 'overview',
          component: () => import('@/views/overview/OverviewView.vue'),
          meta: { title: '概览' },
        },
        {
          path: 'categories',
          name: 'categories',
          component: () => import('@/views/categories/CategoriesView.vue'),
          meta: { title: '分类管理' },
        },
        {
          path: 'posts',
          name: 'posts',
          component: () => import('@/views/posts/PostsView.vue'),
          meta: { title: '帖子管理' },
        },
        {
          path: 'comments',
          name: 'comments',
          component: () => import('@/views/comments/CommentsView.vue'),
          meta: { title: '评论审核' },
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: { name: 'overview' } },
  ],
});

router.afterEach((to) => {
  const pageTitle = to.meta.title;
  document.title = pageTitle
    ? `${String(pageTitle)} | ${APP_TITLE}`
    : APP_TITLE;
});

export default router;
