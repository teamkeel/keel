import { WithQueryParams, permissions } from "@teamkeel/sdk";

export default WithQueryParams(async (ctx, inputs) => {
  permissions.allow();

  return inputs;
});
