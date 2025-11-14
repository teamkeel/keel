import { FileInput, FlowConfig, models, File } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default FileInput(config, async (ctx) => {
  const { avatar, passport } = await ctx.ui.page("file input page", {
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

  const { userId } = await ctx.step("create user", async () => {
    // checking that the returned data from the page step (avatar) is a file
    if (!(avatar instanceof File)) {
      throw new Error("not a file");
    }

    const url = await avatar.getPresignedUrl();
    const user = await models.user.create({
      avatar: avatar,
      passport: passport,
    });

    return {
      userId: user.id,
      avatarUrl: url.toString(),
    };
  });

  const imageUrl = await ctx.step("get image", async () => {
    const user = await models.user.findOne({ id: userId });
    if (!user) {
      throw new Error("user not found");
    }

    const url = await user.avatar.getPresignedUrl();
    return url.toString();
  });

  return {
    id: userId,
    avatarUrl: imageUrl,
  };
});
