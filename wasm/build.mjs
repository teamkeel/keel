import { build } from 'esbuild';

import watPlugin from 'esbuild-plugin-wat'

build({
  outdir: 'dist',
  bundle: true,
  platform: 'node',
  entryPoints: ['index.ts'],
  plugins: [
    watPlugin()
  ]
})