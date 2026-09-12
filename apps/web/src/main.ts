import { createApp } from 'vue';

import App from './App.vue';
import router from './router';
import './styles.css';
import { useCategoryStore } from './stores/category';
import { pinia } from './stores';
import { webKit } from './web-kit';

const app = createApp(App);

app.use(webKit).use(pinia);
await useCategoryStore().initialize();
app.use(router).mount('#app');
