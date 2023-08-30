import { VerifyEmail, models } from "@teamkeel/sdk";

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

export default VerifyEmail(async (_, event) => {
  switch (event.eventName) {
    case "member.created":
      await sleep(1000); // Tests that jobs are being awaited correctly in tests
      await models.member.update({ id: event.target.id }, { verified: true });
    default:
      break;
  }
});
