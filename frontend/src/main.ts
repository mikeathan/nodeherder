import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import Notifications from '@kyvg/vue3-notification';
import '@fortawesome/fontawesome-free/css/all.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap';
import { store, key } from './store/index';
import mitt from 'mitt';
import { Events } from '@/types/events.type';
import VueApexCharts from 'vue3-apexcharts';
import VueDatePicker from '@vuepic/vue-datepicker';
const emitter = mitt<Events>();

//import "./assets/css/styles.global.css";
import './assets/css/dark.css';
import '@vuepic/vue-datepicker/dist/main.css';

const app = createApp(App);
app.use(store, key);
app.use(router);
app.use(Notifications);
app.use(VueApexCharts);
app.component('VueDatePicker', VueDatePicker);
app.provide('emitter', emitter);
app.mount('#app');
