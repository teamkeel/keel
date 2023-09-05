import { InviteGuest, InviteGuestHooks, InviteStatus, models } from '@teamkeel/sdk';


// To learn more about what you can do with hooks,
// visit https://docs.keel.so/functions
const hooks : InviteGuestHooks = {};

export default InviteGuest({
    afterWrite: async (ctx, inputs, data) => {
        if (data.isFamily ) {
            await models.weddingInvitee.update({ id: data.id }, {status: InviteStatus.Accepted});
        }
    }
});
	