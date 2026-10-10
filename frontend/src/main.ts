import { createApp } from 'vue';
import PrimeVue from 'primevue/config';
import ToastService from 'primevue/toastservice';
import VueApexCharts from 'vue3-apexcharts';

import '@fontsource/figtree/400.css';
import '@fontsource/figtree/500.css';
import '@fontsource/figtree/600.css';
import '@fontsource/figtree/700.css';
import '@fontsource/figtree/800.css';
import '@fontsource/lexend/300.css';
import '@fontsource/lexend/500.css';
import 'primeicons/primeicons.css';
import 'primeflex/primeflex.css';
import '@/assets/styles/tokens.css';
import '@/assets/styles/themes.css';
import '@/assets/styles/base.css';

import App from './App.vue';
import router from './router';
import { store, key } from './store/index';
import ClickOutside from './directives/click-outside';
import { NodeHerderPreset, PRIMEVUE_THEME_OPTIONS } from '@/theme/primevue-preset';
import { startThemeSettings } from '@/composables/useThemeSettings';

// Apply the stored appearance before the first paint so there is no flash of the wrong theme.
startThemeSettings();

const app = createApp(App);

app.use(PrimeVue, { theme: { preset: NodeHerderPreset, options: PRIMEVUE_THEME_OPTIONS } });
app.use(ToastService);
app.directive('click-outside', ClickOutside);
app.use(store, key);
app.use(router);
app.component('ApexChart', VueApexCharts);
app.mount('#app');
