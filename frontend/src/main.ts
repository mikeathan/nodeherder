import { createApp } from 'vue';
import App from './App.vue';

import router from './router';
import '@fortawesome/fontawesome-free/css/all.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap';
import { store, key } from './store/index';
import mitt from 'mitt';
import { Events } from '@/types/events.type';
import VueApexCharts from 'vue3-apexcharts';
import PrimeVue from 'primevue/config'; // here
import "primevue/resources/themes/saga-blue/theme.css"; //theme
import "primevue/resources/primevue.min.css"; //core CSS
import "primeicons/primeicons.css"; //icons

const emitter = mitt<Events>();

const app = createApp(App);
app.use(PrimeVue);
app.use(store, key);
app.use(router);
app.use(VueApexCharts);
app.provide('emitter', emitter);
app.mount('#app');
