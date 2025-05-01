import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["cjs", "esm"],
  dts: {
    entry: "./src/index.d.ts", // ← overrides inferred declaration generation
  },
  splitting: false,
  sourcemap: true,
  keepNames: true,
  clean: true,
  target: "node16",
  loader: {
    ".js": "jsx",
  },
});
