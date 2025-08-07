import { UserDelays } from "@teamkeel/sdk";

export default UserDelays({}, async (ctx) => {
  await ctx.step(
    "user defined delay step",
    {
      retries: 3,
      retryPolicy: (retry: number) => {
        switch (retry) {
          case 1:
            return 3000; // retry 1, delay 3s
          case 2:
            return 1000; // retry 2, delay 1s
          default:
            return 2000; // retry 3, delay 2s
        }
      },
    },
    async (args) => {
      if (args.attempt !== 4) {
        throw new Error("enforce 3 retries");
      }

      return "completed";
    }
  );
});
