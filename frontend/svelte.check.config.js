import baseConfig from './svelte.config.js';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  ...baseConfig,
  compilerOptions: {
    ...baseConfig.compilerOptions,
    runes: true
  }
};

export default config;
