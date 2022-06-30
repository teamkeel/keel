import { build } from 'esbuild';
import watPlugin from 'esbuild-plugin-wat';
import NpmDts from 'npm-dts';

const { Generator } = NpmDts;

build({
  outdir: 'dist',
  bundle: true,
  platform: 'node',
  entryPoints: ['index.ts'],
  plugins: [
    watPlugin()
  ]
});

new Generator({
  entry: 'typings.ts',
  output: 'dist/index.d.ts'
}).generate();
