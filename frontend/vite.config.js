import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
import Components from 'unplugin-vue-components/vite';
import { PrimeVueResolver } from '@primevue/auto-import-resolver';
import { createRequire } from 'module';

const require = createRequire(import.meta.url);
const { getAppVersion } = require('./scripts/get-version.cjs');
const appVersion = process.env.APP_VERSION || getAppVersion();

export default defineConfig({
  plugins: [
    vue(),
    Components({
      resolvers: [PrimeVueResolver()],
    }),
  ],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  server: {
    port: 4100,
    host: '0.0.0.0',
    allowedHosts: true,
  },
  preview: {
    port: 9080,
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
});
