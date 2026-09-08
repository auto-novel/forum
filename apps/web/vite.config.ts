import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';
import { defineConfig, loadEnv, type UserConfig } from 'vite';

function setupAuthProxy(config: UserConfig) {
  const authUrl = 'https://auth.novelia.cc';
  const proxy = config.server!.proxy!;

  proxy['/auth-proxy/api'] = {
    target: authUrl,
    changeOrigin: true,
    rewrite: (path) => path.replace(/^\/auth-proxy/, ''),
  };

  proxy['/auth-proxy/assets'] = {
    target: authUrl,
    changeOrigin: true,
    rewrite: (path) => path.replace(/^\/auth-proxy\/assets/, '/assets'),
  };

  proxy['/auth-proxy'] = {
    target: authUrl,
    changeOrigin: true,
    rewrite: (path) => path.replace(/^\/auth-proxy/, ''),
    selfHandleResponse: true,
    headers: { 'accept-encoding': 'identity' },
    configure(proxyServer) {
      proxyServer.on('proxyRes', (proxyResponse, _request, response) => {
        const chunks: Buffer[] = [];
        proxyResponse.on('data', (chunk: Buffer) => chunks.push(chunk));
        proxyResponse.on('end', () => {
          const body = Buffer.concat(chunks)
            .toString()
            .replaceAll('/assets', '/auth-proxy/assets');

          response.statusCode = proxyResponse.statusCode ?? 200;
          response.statusMessage = proxyResponse.statusMessage ?? '';
          for (const [key, value] of Object.entries(proxyResponse.headers)) {
            if (value !== undefined) response.setHeader(key, value);
          }
          response.removeHeader('content-length');
          response.end(body);
        });
      });
    },
  };
}

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(
    mode,
    fileURLToPath(new URL('.', import.meta.url)),
    'VITE_',
  );
  const isServe = command === 'serve';
  const apiMode = env.VITE_API_MODE;
  const apiUrl =
    apiMode === 'native'
      ? 'http://localhost:8080'
      : apiMode === 'local'
        ? 'http://localhost:5000'
        : 'https://forum.novelia.cc';

  const config: UserConfig = {
    define: {
      __AUTH_URL__: JSON.stringify(
        isServe ? '/auth-proxy/' : 'https://auth.novelia.cc',
      ),
    },
    plugins: [vue(), tailwindcss()],
    optimizeDeps: {
      exclude: ['@novelia/auth-api'],
    },
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

  if (isServe) setupAuthProxy(config);

  return config;
});
