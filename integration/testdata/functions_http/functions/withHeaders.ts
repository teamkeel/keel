import { WithHeaders, permissions } from "@teamkeel/sdk";

export default WithHeaders(async (ctx, inputs) => {
  permissions.allow();

  const value = ctx.headers.get("X-MyRequestHeader");
  ctx.response.headers.set("X-MyResponseHeader", value || "");

  return {};
});
