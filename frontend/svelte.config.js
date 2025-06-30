import adapter from '@sveltejs/adapter-cloudflare';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      // Workers-specific adapter configuration
      routes: {
        include: ['/*'],
        exclude: ['<all>']
      }
    }),
    alias: {
      '@/*': './src/lib/*'
    }
  }
};

export default config;
