import { CustomRead } from "@teamkeel/sdk";

export default CustomRead(async (ctx, inputs) => {
  return {
    message: "This is a custom read action",
    timestamp: new Date().toISOString(),
  };
});
