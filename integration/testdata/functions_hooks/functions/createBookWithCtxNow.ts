import { CreateBookWithCtxNow } from "@teamkeel/sdk";

// Test: ctx.now() returns the current timestamp
export default CreateBookWithCtxNow({
  beforeWrite(ctx, inputs, values) {
    // Use ctx.now() to get the current timestamp
    const now = ctx.now();
    return {
      ...values,
      createdAtFromCtx: now,
    };
  },
});
