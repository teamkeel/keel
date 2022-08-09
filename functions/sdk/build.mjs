import { build } from 'esbuild';
// import NpmDts from 'npm-dts';

// const { Generator } = NpmDts;

build({
  outdir: 'dist',
  platform: 'node',
  entryPoints: [
    'index.ts'
  ],
  bundle: true,
  external: [
    '../client'
  ],
  tsconfig: 'tsconfig.build.json'
});

// Doesnt generate absolutely correct types due to several annoying issues but renable
// to get a good basis
// new Generator({
//   entry: './index.ts',
//   output: 'dist/index.d.ts'
// }, true, true).generate();
