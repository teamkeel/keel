import { models, UpdateHeadCount, InviteStatus } from "@teamkeel/sdk";

// To learn more about jobs, visit https://docs.keel.so/jobs
export default UpdateHeadCount(async (ctx, inputs) => {
  const guests = await models.weddingInvitee.findMany();
  let count = 0;
  let shouldErr = false;

  for (var guest of guests) {
    if (guest.firstName == "Prisma") {
      shouldErr = true;
    }

    if (guest.status == InviteStatus.Declined) {
      await models.weddingInvitee.delete({ id: guest.id });
    } else if (guest.status == InviteStatus.Accepted) {
      count++;
    }
  }

  await models.wedding.update({ id: inputs.weddingId }, { headcount: count });

  if (shouldErr) {
    // This will __NOT__ rollback the function, mutations nor logs
    throw new Error("prisma is not invited!");
  }
});
