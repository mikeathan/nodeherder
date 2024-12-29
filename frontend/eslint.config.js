import vue from 'eslint-plugin-vue';
import vueParser from 'vue-eslint-parser';

export default [
  {
    ignores: ['dist'], // Exclude the 'dist' folder
  },
  {
    plugins: {
      vue,
    },
    languageOptions: {
      parser: vueParser,
      parserOptions: {
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
    },
    rules: {
      ...vue.configs.recommended.rules, // Vue recommended rules
      'no-unused-vars': 'warn',
      'no-console': 'warn',
      'indent': ['error', 2],
      'quotes': ['error', 'single'],
      'semi': ['error', 'always'],
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
