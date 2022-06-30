import { build } from 'esbuild';
import watPlugin from 'esbuild-plugin-wat';
// import NpmDts from 'npm-dts';

// const { Generator } = NpmDts;

build({
  outdir: 'dist',
  bundle: true,
  platform: 'node',
  entryPoints: ['index.ts'],
  plugins: [
    watPlugin()
  ]
});

// Todo: broken typings generated with ambient relative import
// for the meantime add typings manually
// new Generator({
//   entry: 'index.ts',
//   output: 'dist/index.d.ts'
// }).generate();
