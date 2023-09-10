import {
  models,
  UpdateHeadCountWithKysely,
  InviteStatus,
  useDatabase,
} from "@teamkeel/sdk";

// To learn more about jobs, visit https://docs.keel.so/jobs
export default UpdateHeadCountWithKysely(async (ctx, inputs) => {
  const guests = await models.weddingInvitee.findMany();
  let count = 0;

  for (var guest of guests) {
    if (guest.status == InviteStatus.Declined) {
      await useDatabase()
        .deleteFrom("wedding_invitee")
        .where("id", "=", guest.id)
        .execute();
    } else if (guest.status == InviteStatus.Accepted) {
      count++;
    }
  }

  await useDatabase()
    .updateTable("wedding")
    .set({ headcount: count })
    .where("id", "=", inputs.weddingId)
    .execute();
});
