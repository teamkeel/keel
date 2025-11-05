import { TextInput, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default TextInput(config, async (ctx) => {
  const page1 = await ctx.ui.page("text input page", {
    content: [
      ctx.ui.inputs.text("username", {
        label: "Username",
        defaultValue: "guest",
        validate: (data) => {
          // Username must be at least 3 characters
          if (data.length < 3) {
            return "Username must be at least 3 characters";
          }
          // Username cannot contain spaces
          if (data.includes(" ")) {
            return "Username cannot contain spaces";
          }
          return true;
        },
      }),
      ctx.ui.inputs.text("email", {
        label: "Email Address",
        optional: true,
        validate: (data) => {
          // Only validate if email is provided (since it's optional)
          if (data && !data.includes("@")) {
            return "Email must contain @";
          }
          return true;
        },
      }),
      ctx.ui.inputs.text("description", {
        label: "Description",
        helpText: "Enter a brief description",
        validate: (data) => {
          // Description must be between 5 and 100 characters
          if (data.length < 5) {
            return "Description must be at least 5 characters";
          }
          if (data.length > 100) {
            return "Description must be at most 100 characters";
          }
          return true;
        },
      }),
    ],
  });

  return {
    username: page1.username,
    email: page1.email,
    description: page1.description,
  };
});
