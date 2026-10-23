import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
import Components from 'unplugin-vue-components/vite';
import { PrimeVueResolver } from '@primevue/auto-import-resolver';
import { createRequire } from 'module';

const require = createRequire(import.meta.url);
const { getAppVersion } = require('./scripts/get-version.cjs');
const appVersion = process.env.APP_VERSION || getAppVersion();

export default defineConfig(({ mode }) => ({
  plugins: [
    vue(),
    Components({
      resolvers: [PrimeVueResolver()],
    }),
  ],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  /*
   * Browser tests load routes lazily, so the dependency optimizer would otherwise discover a
   * route-only dependency (for example primevue/textarea for /assistant) mid-test and force a
   * full page reload that keeps the previous URL. Scan the component graph up front in e2e mode
   * so optimization happens before any navigation. Development and production are unchanged.
   */
  server: {
    port: 4100,
    host: '0.0.0.0',
    allowedHosts: true,
    // Warm every route component so dependency optimization finishes before tests navigate.
    warmup: mode === 'e2e' ? { clientFiles: ['./src/router/index.ts', './src/components/**/*.vue'] } : undefined,
  },
  preview: {
    port: 9080,
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
}));
