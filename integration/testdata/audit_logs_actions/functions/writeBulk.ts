import { InviteMany, models } from '@teamkeel/sdk';

export default InviteMany(async (ctx, inputs) => {
    for (var name of inputs.names) {
        if (name == "Prisma") {
            throw new Error("prisma is not invited!")
        }

        await models.weddingInvitee.create({
            firstName: name,
            weddingId: inputs.weddingId
        })
    }

    return true;

})