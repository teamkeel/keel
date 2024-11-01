import { PresignedUrl } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default PresignedUrl(async (ctx, inputs) => {
  const file = await inputs.file.store();

  const url = await file.getPresignedUrl();

  return url;
});
