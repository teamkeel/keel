// This file is loaded by Vitest because of the ../vitest.config.mjs file
// which specifies it as a setupFile. When running tests with `keel test`
// we tell Vitest to load that config file.

import { expect } from "vitest";
import { toHaveError } from "./toHaveError.mjs";
import { toHaveAuthorizationError } from "./toHaveAuthorizationError.mjs";

expect.extend({
  toHaveError,
  toHaveAuthorizationError,
});
