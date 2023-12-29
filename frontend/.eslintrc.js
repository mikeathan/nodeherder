module.exports = {
  // Use the rule set.
  extends: ["plugin:vue/base"],
  rules: {
    // Enable vue/script-setup-uses-vars rule
    "vue/script-setup-uses-vars": "error",
  },
  parserOptions: {
    parser: "@typescript-eslint/parser",
  },
  plugins: ["@typescript-eslint"],
};
