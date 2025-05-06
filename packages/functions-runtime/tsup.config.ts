import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["cjs", "esm"],
  dts: true,
  splitting: false,
  sourcemap: true,
  keepNames: true,
  clean: true,
  target: "node22",
  loader: {
    ".js": "jsx",
  },
});
