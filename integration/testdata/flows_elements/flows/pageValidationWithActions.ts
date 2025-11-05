import { PageValidationWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default PageValidationWithActions(config, async (ctx) => {
  const page1 = await ctx.ui.page("page level validation", {
    content: [
      ctx.ui.inputs.text("name", {
        label: "Name",
      }),
      ctx.ui.inputs.number("age", {
        label: "Age",
      }),
    ],
    actions: ["submit", "draft"],
    validate: (data, action) => {
      // Verify action parameter is passed correctly when an action is provided
      if (action !== undefined && action !== "submit" && action !== "draft") {
        throw new Error(
          `Expected action to be 'submit' or 'draft', got: ${action}`
        );
      }

      // When submitting, name is required
      if (action === "submit" && !data.name) {
        return "Name is required when submitting";
      }
      // When submitting, age must be >= 18
      if (action === "submit" && data.age < 18) {
        return "Must be 18 or older to submit";
      }
      // Draft allows any values
      return true;
    },
  });

  return {
    action: page1.action,
    name: page1.data.name,
    age: page1.data.age,
  };
});
