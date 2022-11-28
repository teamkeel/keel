import { build } from "esbuild";
import NpmDts from "npm-dts";

const { Generator } = NpmDts;

build({
  outdir: "dist",
  platform: "node",
  entryPoints: ["index.ts"],
  bundle: true,
  tsconfig: "tsconfig.build.json",
});

// Generate an index.d.ts file automatically based on the typescript source code.
new Generator(
  {
    entry: "./index.ts",
    output: "dist/index.d.ts",
    // use the build version of the tsconfig in order to exclude some test files
    tsc: "-p tsconfig.build.json",
  },
  true,
  true
).generate();
