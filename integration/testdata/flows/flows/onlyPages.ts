import { OnlyPages } from "@teamkeel/sdk";

// To learn more about flows, visit https://docs.keel.so/flows
export default OnlyPages({}, async (ctx, inputs) => {
  const grid = ctx.ui.display.grid({
    data: [{ a: "A thing" }],
    render: (d) => ({
      title: d.a,
    }),
  });

  await ctx.ui.page("first page", {
    title: "Grid of things",
    content: [grid],
  });

  await ctx.ui.page("question", {
    title: "My flow",
    content: [
      ctx.ui.inputs.boolean("yesno", {
        label: "Did you like the things?",
      }),
    ],
  });
});
