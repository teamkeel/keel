import { VerifyEmail, models } from "@teamkeel/sdk";

export default VerifyEmail(async (_, event) => {
  switch (event.eventName) {
    case "member.created":
      await models.member.update({ id: event.target.id }, { verified: true });
    default:
      break;
  }
});
