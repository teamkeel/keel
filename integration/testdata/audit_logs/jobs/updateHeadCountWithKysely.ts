import { models, UpdateHeadCountWithKysely, WeddingInvitee, InviteStatus, useDatabase } from "@teamkeel/sdk";


// To learn more about jobs, visit https://docs.keel.so/jobs
export default UpdateHeadCountWithKysely(async (ctx, inputs) => {
  const guests = await models.weddingInvitee.findMany();
  let count = 0;
  let shouldErr = false;

  for (var guest of guests) {
    if (guest.firstName == "Prisma") {
      shouldErr = true;
    }

    if (guest.status == InviteStatus.Declined) {
     // await models.weddingInvitee.delete({ id: guest.id });

      await useDatabase()
        .deleteFrom("wedding_invitee")
        .where("id", "=", guest.id)
        .execute();

    } else if (guest.status == InviteStatus.Accepted) {
      count++;
    }
  }

 // await models.wedding.update({ id: inputs.weddingId }, { headcount: count });
 await useDatabase()
  .updateTable("wedding")
  .set({ headcount: count })
  .where("id", "=", inputs.weddingId)
  .execute();
});
