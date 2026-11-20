import { defineConfig } from 'vite-plus';

export default defineConfig({
  staged: {
    '*': 'vp check --fix',
  },
  fmt: {
    singleQuote: true,
    printWidth: 80,
  },
  lint: {
    ignorePatterns: ['**/*.d.ts', '**/coverage/**', '**/dist/**', 'data/**'],
    plugins: ['typescript', 'unicorn', 'oxc', 'vue'],
    categories: { correctness: 'error' },
    rules: {
      'prefer-const': 'warn',
      'typescript/consistent-type-imports': 'error',
    },
  },
  run: {
    cache: true,
  },
});
