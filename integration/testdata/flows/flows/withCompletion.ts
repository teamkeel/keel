import { WithCompletion } from "@teamkeel/sdk";


export default WithCompletion({ stages: [
  {
    key: "starting",
    name: "Starting",
    description: "this is the starting stage",
  },
  {
    key: "ending",
    name: "Ending",
    description: "this is the ending stage",
  },
], }, async (ctx) => {
  await ctx.step("my step", { stage: "starting" }, async () => {
    return;
  });

  return ctx.complete({
    title: "Completed flow",
    description: "this complete page replaces the normal end page",
    stage: "ending",
    content: [ctx.ui.display.markdown({content:"congratulations"})],
    data: {
      value: "flow value",
    },
  });
});
