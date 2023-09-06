import { InviteMany, models } from "@teamkeel/sdk";

export default InviteMany(async (ctx, inputs) => {
  let shouldErr = false;

  for (var name of inputs.names) {
    if (name == "Prisma") {
      shouldErr = true;
    }

    await models.weddingInvitee.create({
      firstName: name,
      weddingId: inputs.weddingId,
    });
  }

  if (shouldErr) {
    // This will rollback the function, any mutations and any audit logs
    throw new Error("prisma is not invited!");
  }

  return true;
});
