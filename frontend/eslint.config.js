import tsPlugin from '@typescript-eslint/eslint-plugin';
import tsParser from '@typescript-eslint/parser';
import vue from 'eslint-plugin-vue';
import vueParser from 'vue-eslint-parser';

const ecmaVersion = 2022;
const globals = {
  window: 'readonly',
  document: 'readonly',
  console: 'readonly',
  process: 'readonly',
};
const parserOptions = {
  ecmaVersion,
  sourceType: 'module',
  ecmaFeatures: {
    jsx: true, // Enable JSX support if needed
  },
};

export default [
  {
    ignores: ['dist', 'playwright-report', 'test-results'], // Generated build and browser-test artifacts
  },
  {
    files: ['**/*.{js,cjs,mjs}'],
    plugins: {
      vue,
    },
    languageOptions: {
      ecmaVersion,
      parserOptions,
      globals,
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
    // Vue single-file components were previously skipped: the old `files: ['*.vue']`
    // pattern only matched the project root, and the removed
    // `vue/script-setup-uses-vars` rule aborted a root-level run. Parse
    // <script setup lang="ts"> with the TypeScript parser and keep only
    // correctness rules here; stylistic rules stay on plain JS files.
    files: ['**/*.vue'],
    plugins: {
      '@typescript-eslint': tsPlugin,
      vue,
    },
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        ...parserOptions,
        parser: tsParser,
      },
      globals,
    },
    settings: {
      vue: {
        version: 'detect', // Automatically detect Vue version
      },
    },
    rules: {
      // Essential correctness rules only: the recommended preset also adds
      // formatting-order rules (attributes/block order), which would fail on
      // unrelated components and belong in a dedicated formatting change.
      ...vue.configs['essential'].rules,
      '@typescript-eslint/no-unused-vars': 'warn',
      'no-console': 'warn',
      'vue/multi-word-component-names': 'off', // Disable multi-word component names warning
    },
  },
];
