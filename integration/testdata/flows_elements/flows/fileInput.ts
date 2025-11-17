import { FileInput, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default FileInput(config, async (ctx) => {
  const page1 = await ctx.ui.page("file input page", {
    content: [
      ctx.ui.inputs.file("avatar", {
        label: "Avatar",
        optional: true,
        helpText: "A nice photo of yourself",
      }),
      ctx.ui.inputs.file("passport", {
        label: "Passport",
      }),
    ],
  });

  return {
    avatar: page1.avatar,
    passport: page1.passport,
  };
});
