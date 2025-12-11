import { ChildTask } from "@teamkeel/sdk";

export default ChildTask({}, async (ctx, inputs) => {
  // Simple flow that just completes with the input data
  return ctx.complete({
    data: {
      parentName: inputs.parentName,
      index: inputs.index,
    },
  });
});
