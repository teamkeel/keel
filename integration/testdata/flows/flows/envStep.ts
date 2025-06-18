import { EnvStep } from "@teamkeel/sdk";

export default EnvStep({}, async (ctx) => {
  await ctx.step("env step", async () => {
    return ctx.env.PERSON_NAME;
  });

  await ctx.step("identity step", async () => {
    return ctx.identity.email;
  });
});
