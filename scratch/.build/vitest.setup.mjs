
import { expect } from "vitest";
import { toHaveError, toHaveAuthorizationError } from "@teamkeel/testing-runtime";

expect.extend({
	toHaveError,
	toHaveAuthorizationError,
});
			