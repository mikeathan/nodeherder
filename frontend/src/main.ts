import { createApp } from 'vue';
import App from './App.vue';

import router from './router';
import '@fortawesome/fontawesome-free/css/all.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap';
import '@fontsource/roboto';

import { store, key } from './store/index';
import mitt from 'mitt';
import { Events } from '@/types/events.type';
import VueApexCharts from 'vue3-apexcharts';

import PrimeVue from 'primevue/config';
import Aura from '@primevue/themes/aura';
import Material from '@primevue/themes/material';

import ToastService from 'primevue/toastservice';
import { MaterialBlue } from './themes/material_blue.js';
import 'primeicons/primeicons.css'; // Icons
import 'primeflex/primeflex.css';

const emitter = mitt<Events>();

const app = createApp(App);

app.use(PrimeVue, {
  theme: {
    preset: MaterialBlue,
    options: {
      darkModeSelector: 'system',
    },
  },
});

app.use(ToastService);
app.use(store, key);
app.use(router);
app.use(VueApexCharts);
app.provide('emitter', emitter);
app.mount('#app');
