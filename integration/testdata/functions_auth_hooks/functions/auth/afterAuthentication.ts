import { AfterAuthentication, models } from "@teamkeel/sdk";

// This synchronous hook will execute after authentication has complete
export default AfterAuthentication(async (ctx) => {
  if (ctx.env.TEST != "test") {
    throw new Error("expected ctx.env.TEST to be set to 'test'");
  }

  if (ctx.isAuthenticated) {
    if (!ctx.identity) {
      throw new Error("ctx.identity must not be empty");
    }

    const account = await models.account
      .where({ identityId: ctx.identity!.id })
      .findOne();
    if (account) {
      const newCount = account.loginCount + 1;
      const updated = await models.account.update(
        { id: account.id },
        { loginCount: newCount }
      );
    }
  } else {
    if (ctx.identity) {
      throw new Error("ctx.identity must be empty");
    }
  }
});
