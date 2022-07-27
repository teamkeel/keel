import { build } from 'esbuild';
import NpmDts from 'npm-dts';

const { Generator } = NpmDts;

build({
  outdir: 'dist',
  bundle: true,
  platform: 'node',
  allowOverwrite: true,
  entryPoints: ['index.ts']
});

new Generator({
  entry: 'index.ts',
  output: 'dist/index.d.ts',
  force: true
}, true, true).generate();
