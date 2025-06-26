import { WithCompletionMinimal } from "@teamkeel/sdk";

export default WithCompletionMinimal({}, async (ctx) => {
  return ctx.complete({
    title: "Completed flow",
  });
});
