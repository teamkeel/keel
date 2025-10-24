import { CustomWrite } from "@teamkeel/sdk";

export default CustomWrite(async (ctx, inputs) => {
  return {
    success: true,
    message: "This is a custom write action",
    timestamp: new Date().toISOString(),
  };
});
