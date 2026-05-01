import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import { transform } from 'esbuild';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [
      {
        name: 'treat-js-files-as-jsx',
        async transform(code, id) {
          if (!id.includes('/src/') || !id.endsWith('.js')) {
            return null;
          }

          const result = await transform(code, {
            loader: 'jsx',
            jsx: 'automatic',
            sourcemap: false,
          });

          return {
            code: result.code,
            map: null,
          };
        },
      },
      react({
        include: /\.[jt]sx?$/,
      }),
    ],
    esbuild: {
      loader: 'jsx',
      include: /src\/.*\.js$/,
    },
    optimizeDeps: {
      esbuildOptions: {
        loader: {
          '.js': 'jsx',
        },
      },
    },
    server: {
      host: '0.0.0.0',
      port: 3000,
    },
    preview: {
      host: '0.0.0.0',
      port: 3000,
    },
    define: {
      __API_URL__: JSON.stringify(env.VITE_API_URL || 'http://localhost:8080'),
    },
    test: {
      globals: true,
      environment: 'jsdom',
      setupFiles: ['./src/setupTests.js'],
      alias: {
        'react-json-tree': new URL('./src/__mocks__/react-json-tree.js', import.meta.url).pathname,
      },
    },
  };
});
