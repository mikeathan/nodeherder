import { createApp } from 'vue';
import App from './App.vue';

import 'vuetify/styles';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import router from './router';
import '@fortawesome/fontawesome-free/css/all.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap';
import { store, key } from './store/index';
import mitt from 'mitt';
import { Events } from '@/types/events.type';
import VueApexCharts from 'vue3-apexcharts';

import 'vuetify/styles';

const emitter = mitt<Events>();

//import "./assets/css/styles.global.css";
import './assets/css/dark.css';
const vuetify = createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'dark',
  },
});

const app = createApp(App);
app.use(vuetify);
app.use(store, key);
app.use(router);
app.use(VueApexCharts);
app.provide('emitter', emitter);
app.mount('#app');
