import { SubscriberWithException, models } from "@teamkeel/sdk";

// To learn more about events and subscribers, visit https://docs.keel.so/events
export default SubscriberWithException(async (ctx, event) => {
  const tracker = (await models.trackSubscriber.findMany())[0];
  await models.trackSubscriber.update(
    { id: tracker.id },
    { didSubscriberRun: true }
  );
  throw new Error("something bad has happened!");
});
