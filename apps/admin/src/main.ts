import { createApp } from 'vue';
import App from './App.vue';

import { createAdminAuthGuard, createAdminKit } from '@novelia/admin-kit';

import { createForumApi, forumApiKey } from './api';
import router from './router';

const adminKit = createAdminKit({
  auth: {
    app: 'f',
    url: __AUTH_URL__,
  },
  brand: 'Forum',
  repository: {
    url: 'https://github.com/auto-novel/forum',
    buildTime: __BUILD_TIME__,
    commitSha: __COMMIT_SHA__,
  },
});

router.beforeEach(createAdminAuthGuard(adminKit));

createApp(App)
  .provide(forumApiKey, createForumApi(adminKit.api))
  .use(adminKit)
  .use(router)
  .mount('#app');
