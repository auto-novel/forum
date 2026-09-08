import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  const env = loadEnv(
    mode,
    fileURLToPath(new URL('.', import.meta.url)),
    'VITE_',
  );
  const apiMode = env.VITE_API_MODE;
  const apiUrl =
    apiMode === 'native'
      ? 'http://localhost:8080'
      : apiMode === 'local'
        ? 'http://localhost:5000'
        : 'https://forum.novelia.cc';

  return {
    plugins: [vue(), tailwindcss()],
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: apiUrl,
          changeOrigin: true,
          rewrite:
            apiMode === 'native'
              ? (path) => path.replace(/^\/api/, '')
              : undefined,
        },
      },
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
  };
});
