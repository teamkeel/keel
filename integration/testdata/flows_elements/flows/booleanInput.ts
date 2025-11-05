import { BooleanInput, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default BooleanInput(config, async (ctx) => {
  const page1 = await ctx.ui.page("boolean input page", {
    content: [
      ctx.ui.inputs.boolean("isActive", {
        label: "Is Active",
        defaultValue: true,
        validate: (data) => {
          // isActive must be true
          if (!data) {
            return "Account must be active";
          }
          return true;
        },
      }),
      ctx.ui.inputs.boolean("agreedToTerms", {
        label: "I agree to the terms and conditions",
        optional: true,
        validate: (data) => {
          // If provided, must be true
          if (data === false) {
            return "You must agree to the terms";
          }
          return true;
        },
      }),
      ctx.ui.inputs.boolean("receiveNewsletter", {
        label: "Receive Newsletter",
        helpText: "Subscribe to our weekly newsletter",
        defaultValue: false,
        // No validation - any boolean value is acceptable
      }),
    ],
  });

  return {
    isActive: page1.isActive,
    agreedToTerms: page1.agreedToTerms,
    receiveNewsletter: page1.receiveNewsletter,
  };
});
