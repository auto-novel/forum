import { createApp } from 'vue';
import { PiniaColada } from '@pinia/colada';

import App from './App.vue';
import router from './router';
import './styles.css';
import { useCategoryStore } from './stores/category';
import { pinia } from './stores';
import { webKit } from './web-kit';

const app = createApp(App);

app
  .use(webKit)
  .use(pinia)
  .use(PiniaColada, {
    queryOptions: {
      staleTime: 60_000,
      gcTime: 5 * 60_000,
    },
  });
app.use(router).mount('#app');
void useCategoryStore().initialize();
