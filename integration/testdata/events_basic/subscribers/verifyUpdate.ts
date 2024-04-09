import { models, VerifyUpdate, SubscriberContextAPI } from "@teamkeel/sdk";

export default VerifyUpdate(async (ctx: SubscriberContextAPI, event) => {
  if (event.target.data.name == "") {
    throw new Error("name cannot be empty");
  }

  if (event.target.previousData == null) {
    throw new Error("previous data cannot be null");
  }

  if (!event.target.data.verifiedUpdate) {
    await models.person.update(
      { id: event.target.data.id },
      { verifiedUpdate: true }
    );
  }
});
