import { MultipleActions } from "@teamkeel/sdk";

// To learn more about flows, visit https://docs.keel.so/flows
export default MultipleActions({}, async (ctx, inputs) => {
  const { action } = await ctx.ui.page("question", {
    title: "Continue flow?",
    content: [
      ctx.ui.inputs.boolean("yesno", {
        label: "Did you like the things?",
      }),
    ],
    actions: ["finish", "continue"],
  });

  if (action == "finish") {
    return;
  }

  await ctx.ui.page("another-question", {
    title: "Another question",
    content: [
      ctx.ui.inputs.text("name", {
        label: "Name",
      }),
    ],
    allowBack: true,
  });
});
