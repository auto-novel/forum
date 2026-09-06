import { createRouter, createWebHistory } from 'vue-router';

import { AdminLoginView } from '@novelia/admin-kit';

import OverviewView from '@/views/overview/OverviewView.vue';

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
          component: OverviewView,
          meta: { title: '概览' },
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
