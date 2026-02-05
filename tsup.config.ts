import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["cjs"],
  clean: true,
  noExternal: [/.*/],
  minify: true,
  splitting: false,
  outDir: "dist-sea",
});
