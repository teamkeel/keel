import { build } from 'esbuild';
import NpmDts from 'npm-dts'
import watPlugin from 'esbuild-plugin-wat'

const { Generator } = NpmDts

build({
  outdir: 'dist',
  bundle: true,
  platform: 'node',
  entryPoints: ['index.ts'],
  plugins: [
    watPlugin()
  ]
})

new Generator({
  entry: 'index.ts',
  output: 'dist/index.d.ts'
}).generate()
