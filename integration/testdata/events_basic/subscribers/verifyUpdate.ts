import { models, VerifyUpdate } from "@teamkeel/sdk";

export default VerifyUpdate(async (ctx, event) => {
  switch (event.eventName) {
    case "person.updated":
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

      break;

    case "tracker.updated":
      if (event.target.data.views != event.target.previousData.views + 1) {
        console.log("previous data not correct");
        throw new Error("previous data not correct");
      }

      if (!event.target.data.verifiedUpdate) {
        await models.tracker.update(
          { id: event.target.data.id },
          { verifiedUpdate: true }
        );
      }

      break;
  }
});
