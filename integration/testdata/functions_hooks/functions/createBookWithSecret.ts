import { CreateBookWithSecret } from "@teamkeel/sdk";

// Test: ctx.secrets provides access to secrets defined in keelconfig.yaml
export default CreateBookWithSecret({
  beforeWrite(ctx, inputs, values) {
    // Access the secret from ctx.secrets and store it in the record
    const secretValue = ctx.secrets.TEST_SECRET;
    return {
      ...values,
      createdWithSecret: secretValue || "secret-not-set",
    };
  },
});
