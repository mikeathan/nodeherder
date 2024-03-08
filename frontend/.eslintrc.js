module.exports = {
  // Use the rule set.
  extends: [
    'plugin:vue/base',
    'plugin:prettier/recommended',
    '@vue/prettier',
    '@vue/typescript'
  ],
  rules: {
    // Enable vue/script-setup-uses-vars rule
    'vue/script-setup-uses-vars': 'error'
  },
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser'
  },
  plugins: ['@typescript-eslint', 'vue']
};
