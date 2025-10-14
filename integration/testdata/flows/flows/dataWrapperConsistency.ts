import { DataWrapperConsistency } from "@teamkeel/sdk";

// Test flow to validate data wrapper consistency with and without actions
export default DataWrapperConsistency({}, async (ctx, inputs) => {
  // First page: no actions - should return unwrapped data
  const noActionsResult = await ctx.ui.page("no-actions-page", {
    title: "Page without actions",
    content: [
      ctx.ui.inputs.text("name", {
        label: "Name",
      }),
      ctx.ui.inputs.number("age", {
        label: "Age",
      }),
    ],
  });

  // Second page: with actions - should return wrapped data { data, action }
  const withActionsResult = await ctx.ui.page("with-actions-page", {
    title: "Page with actions",
    content: [
      ctx.ui.inputs.text("city", {
        label: "City",
      }),
    ],
    actions: ["next", "skip"],
  });

  // Return both results for testing
  return ctx.complete({
    data: {
      noActionsResult,
      withActionsResult,
    },
  });
});
