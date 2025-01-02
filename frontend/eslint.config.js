import vue from 'eslint-plugin-vue';
import vueParser from 'vue-eslint-parser';
import { createRequire } from 'module';
const require = createRequire(import.meta.url);

const airbnbBase = require('eslint-config-airbnb-base');
export default [
  {
    ignores: ['dist'], // Exclude the 'dist' folder
  },
  {
    plugins: {
      vue,
      ...airbnbBase.plugins,
    },
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        ...airbnbBase.parserOptions,
        ecmaVersion: 2020,
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true, // Enable JSX support if needed
        },
      },
      globals: {
        window: 'readonly',
        document: 'readonly',
        console: 'readonly',
        process: 'readonly',
      },
    },
    settings: {
      vue: {
        version: 'detect', // Automatically detect Vue version
      },
      ...airbnbBase.settings,
    },

    rules: {
      ...vue.configs.recommended.rules, // Vue recommended rules
      ...airbnbBase.rules,
      'no-unused-vars': 'warn',
      'no-console': 'warn',
      'indent': ['error', 2],
      'quotes': ['error', 'single'],
      'semi': ['error', 'always'],
      'import/no-unresolved': 'off',
    },
  },
  {
    files: ['*.vue'], // Apply specific rules to .vue files
    rules: {
      'vue/script-setup-uses-vars': 'error', // Prevent unused variable warnings in <script setup>
      'vue/multi-word-component-names': 'off', // Disable multi-word component names warning
    },
  },
];
