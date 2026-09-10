import { createApp } from 'vue';

import App from './App.vue';
import { authApi } from './api';
import router from './router';
import './styles.css';
import { initializeTheme } from './theme';

initializeTheme();

const app = createApp(App);

app.onUnmount(authApi.dispose);
app.use(router).mount('#app');
