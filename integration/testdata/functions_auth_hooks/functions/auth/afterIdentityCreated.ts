import { AfterIdentityCreated, models } from '@teamkeel/sdk';

// This synchronous hook will execute after successful authentication and a new identity record created
export default AfterIdentityCreated(async (ctx) => {
    if(ctx.env.TEST != "test") {
        throw new Error("expected ctx.env.TEST to be set to 'test'");
    }

    const identity = ctx.identity!;

    const account = await models.account.findOne({ identityId: identity.id});
    if (!account) {
        await models.account.create({ name: identity.name ?? "Not set", identityId: identity.id });
    } 
});