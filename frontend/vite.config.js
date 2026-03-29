import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
import Components from 'unplugin-vue-components/vite';
import { PrimeVueResolver } from '@primevue/auto-import-resolver';
import { version } from './package.json';
import { execSync } from 'child_process';

let appVersion = version;
try {
  // Grab the latest git tag and strip the leading 'v'.
  const gitVersion = execSync('git describe --tags --abbrev=0').toString().trim().replace(/^v/, '');
  appVersion = gitVersion;
} catch {
  // Silent fallback to package.json if git command fails
}

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
