import { createApp } from 'vue';

import App from './App.vue';
import router from './router';
import './styles.css';
import { webKit } from './web-kit';

const app = createApp(App);

app.use(webKit).use(router).mount('#app');
