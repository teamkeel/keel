import { SubscriberEnvvars, models } from "@teamkeel/sdk";

// To learn more about events and subscribers, visit https://docs.keel.so/events
export default SubscriberEnvvars(async (ctx, _) => {
  if (ctx.env.MY_VAR! != "20") {
    throw new Error("expected env var");
  }

  const tracker = (await models.trackSubscriber.findMany())[0];
  if (!tracker) {
    return;
  }

  await models.trackSubscriber.update(
    { id: tracker.id },
    { didSubscriberRun: true }
  );
});
