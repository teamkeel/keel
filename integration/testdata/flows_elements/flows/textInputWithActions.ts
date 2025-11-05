import { TextInputWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default TextInputWithActions(config, async (ctx) => {
  const page1 = await ctx.ui.page("text input validation", {
    content: [
      ctx.ui.inputs.text("email", {
        label: "Email",
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (action !== undefined && action !== "submit" && action !== "draft") {
            throw new Error(
              `Expected action to be 'submit' or 'draft', got: ${action}`
            );
          }

          // Only validate email format when submitting
          if (action === "submit" && !data.includes("@")) {
            return "Invalid email format";
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
    email: page1.data.email,
  };
});
