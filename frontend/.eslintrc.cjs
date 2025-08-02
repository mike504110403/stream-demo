/* eslint-env node */
require('@rushstack/eslint-patch/modern-module-resolution')

module.exports = {
  root: true,
  extends: [
    'plugin:vue/vue3-essential',
    'eslint:recommended',
    '@vue/eslint-config-typescript',
    '@vue/eslint-config-prettier/skip-formatting'
  ],
  parserOptions: {
    ecmaVersion: 'latest'
  },
  rules: {
    // 關閉一些常見的警告
    'vue/multi-word-component-names': 'off',
    '@typescript-eslint/no-unused-vars': 'warn',
    'no-console': 'off', // 允許 console 語句
    'no-debugger': 'warn',
    'vue/no-mutating-props': 'warn', // 改為警告
    'no-dupe-else-if': 'error' // 保持錯誤
  }
} 