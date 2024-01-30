import { WithForm, permissions } from "@teamkeel/sdk";

export default WithForm(async (ctx, inputs) => {
  permissions.allow();

  return inputs;
});
