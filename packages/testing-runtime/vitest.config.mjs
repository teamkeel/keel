// This config file is used for both running tests in this package
// and also as the config file for Vitest when running `keel test`.

import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    // Using __dirname to get an absolute path for this
    // import as when using with `keel test` a relative
    // import is relative to the current working directory,
    // not this file.
    setupFiles: [__dirname + "/src/vitest-setup"],
  },
});
