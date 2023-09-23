import { models, VerifyEmail, SubscriberContextAPI } from "@teamkeel/sdk";

export default VerifyEmail(async (ctx: SubscriberContextAPI, event) => {
  await models.person.update(
    { id: event.target.data.id },
    { verifiedEmail: true }
  );
});