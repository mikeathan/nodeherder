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

import PrimeVue from 'primevue/config';
import Aura from '@primevue/themes/aura';
import Material from '@primevue/themes/material';
import Button from 'primevue/button';
import InputText from 'primevue/inputtext';
import FloatLabel from 'primevue/floatlabel';
import Menubar from 'primevue/menubar';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import ColumnGroup from 'primevue/columngroup';
import Row from 'primevue/row';
import ToastService from 'primevue/toastservice';
import DatePicker from 'primevue/datepicker';
import Slider from 'primevue/slider';
import InputNumber from 'primevue/inputnumber';
import Select from 'primevue/select';
import ToggleSwitch from 'primevue/toggleswitch';
import Card from 'primevue/card';
import SplitButton from 'primevue/splitbutton';

import DataView from 'primevue/dataview';

import 'primeicons/primeicons.css'; // Icons
import 'primeflex/primeflex.css';

const emitter = mitt<Events>();

const app = createApp(App);
app.use(PrimeVue, {
  theme: {
    preset: Material,
  },
});
app.component('Button', Button);
app.component('InputText', InputText);
app.component('FloatLabel', FloatLabel);
app.component('Menubar', Menubar);
app.component('DataTable', DataTable);
app.component('Column', Column);
app.component('ColumnGroup', ColumnGroup);
app.component('Row', Row);
app.component('DatePicker', DatePicker);
app.component('Slider', Slider);
app.component('InputNumber', InputNumber);
app.component('Select', Select);
app.component('ToggleSwitch', ToggleSwitch);
app.component('Card', Card);
app.component('SplitButton', SplitButton);
app.component('DataView', DataView);
app.use(ToastService);
app.use(store, key);
app.use(router);
app.use(VueApexCharts);
app.provide('emitter', emitter);
app.mount('#app');
