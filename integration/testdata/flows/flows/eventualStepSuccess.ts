import { EventualStepSuccess } from "@teamkeel/sdk";

export default EventualStepSuccess({}, async (ctx) => {
  await ctx.step("erroring step", { retries: 4 }, async (args) => {
    if (args.attempt < 4) {
      throw new Error("Error at attempt " + args.attempt + " of " + args.stepOptions.retries);
    }

    return "Success at attempt " + args.attempt;
  });
});
