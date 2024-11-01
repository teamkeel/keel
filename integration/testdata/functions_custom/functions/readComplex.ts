import { ReadComplex, permissions } from "@teamkeel/sdk";

export default ReadComplex(async (ctx, inputs) => {
  permissions.allow();
  return inputs;
});
