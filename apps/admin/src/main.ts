import { createApp } from 'vue';
import ui from '@nuxt/ui/vue-plugin';
import App from './App.vue';

import { createWebKit } from '@novelia/web-kit';
import './style.css';

import { createForumApi, forumApiKey } from './api';
import router from './router';

const webKit = createWebKit({
  auth: {
    app: 'f',
    url: __AUTH_URL__,
  },
  brand: 'Forum',
  strikes: { enabled: false },
  repository: {
    url: 'https://github.com/auto-novel/forum',
    buildTime: __BUILD_TIME__,
    commitSha: __COMMIT_SHA__,
  },
});

router.beforeEach(async () => {
  await webKit.checkSignedIn();
});

createApp(App)
  .provide(
    forumApiKey,
    createForumApi(
      webKit.createClient(
        new URL('/api/v1/', window.location.origin).toString(),
      ),
    ),
  )
  .use(ui)
  .use(webKit)
  .use(router)
  .mount('#app');
