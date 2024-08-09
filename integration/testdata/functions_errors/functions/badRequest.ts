import { BadRequest, errors } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default BadRequest(async (ctx, inputs) => {
  throw new errors.BadRequest("invalid inputs");

  return;
});
