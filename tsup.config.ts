import { defineConfig } from 'tsup';

export default defineConfig({
  entry: ['src/index.ts'],
  format: ['cjs'],
  dts: false,
  splitting: false,
  sourcemap: false,
  clean: true,
  minify: false,
  target: 'node18',
  outDir: 'dist',
  shims: true,
});