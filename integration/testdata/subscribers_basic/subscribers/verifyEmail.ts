import { VerifyEmail, models } from "@teamkeel/sdk";

export default VerifyEmail(async (_, event) => {
  switch (event.name) {
    case "member.create":
      const updatedMember = await models.member.update(
        { id: event.sourceId },
        { verified: true }
      );
    default:
      break;
  }
});
