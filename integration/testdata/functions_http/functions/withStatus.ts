import { WithStatus, permissions } from "@teamkeel/sdk";
export default WithStatus(async (ctx, inputs) => {
  permissions.allow();

  const { response } = ctx;
  response.headers.set("Location", "https://some.url");
  response.status = 301;

  return null;
});
