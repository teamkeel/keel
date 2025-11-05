import { BooleanInputWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default BooleanInputWithActions(config, async (ctx) => {
  const page1 = await ctx.ui.page("boolean input validation", {
    content: [
      ctx.ui.inputs.boolean("agreeToTerms", {
        label: "I agree to the terms and conditions",
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (action !== undefined && action !== "submit" && action !== "draft") {
            throw new Error(
              `Expected action to be 'submit' or 'draft', got: ${action}`
            );
          }

          // Must agree when submitting
          if (action === "submit" && !data) {
            return "You must agree to the terms to submit";
          }
          // Draft allows any value
          return true;
        },
      }),
    ],
    actions: ["submit", "draft"],
  });

  return {
    action: page1.action,
    agreeToTerms: page1.data.agreeToTerms,
  };
});
