import { build } from 'esbuild';
import NpmDts from 'npm-dts';

const { Generator } = NpmDts;

build({
  outdir: 'dist',
  bundle: true,
  platform: 'node',
  allowOverwrite: true,
  entryPoints: [
    'src/index.ts',
    'src/analyse.ts'
  ],
  external: [
    '../client'
  ]
});

new Generator({
  entry: 'src/index.ts',
  output: 'dist/index.d.ts',
  force: true
}, true, true).generate();
