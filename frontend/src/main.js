import { createApp } from 'vue';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import App from './pages/Project/App.vue';
import router from './router';

createApp(App).use(router).use(ElementPlus).mount('#app');
