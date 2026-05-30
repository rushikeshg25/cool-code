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
  // Ink depends on yoga-layout-prebuilt (asm.js) which esbuild cannot bundle;
  // keep the React/Ink stack external and load it from node_modules at runtime
  // (these are published runtime dependencies).
  external: ['react', 'ink', 'ink-spinner'],
});