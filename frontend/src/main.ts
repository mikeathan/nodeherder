import { createApp } from 'vue';
import App from './App.vue';

import router from './router';
import '@fontsource/roboto';

import { store, key } from './store/index';
import mitt from 'mitt';
import { Events } from '@/types/events.type';
import VueApexCharts from 'vue3-apexcharts';
import ClickOutside from './directives//click-outside';
import PrimeVue from 'primevue/config';
import Aura from '@primevue/themes/aura';
import Material from '@primevue/themes/material';

import ToastService from 'primevue/toastservice';
import { MaterialBlue } from './themes/material_blue.js';
import 'primeicons/primeicons.css'; // Icons
import 'primeflex/primeflex.css';
import '@/assets/styles/variables.css';

const emitter = mitt<Events>();

const app = createApp(App);

app.use(PrimeVue, {
  theme: {
    preset: MaterialBlue,
    options: {
      darkModeSelector: '.dark',
    },
  },
});

app.directive('click-outside', ClickOutside);
app.use(ToastService);
app.use(store, key);
app.use(router);
app.component('ApexChart', VueApexCharts);
app.provide('emitter', emitter);
app.mount('#app');
