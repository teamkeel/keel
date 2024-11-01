import { WriteAccounts, models } from "@teamkeel/sdk";

// To learn more about jobs, visit https://docs.keel.so/jobs
export default WriteAccounts(async (ctx, inputs) => {
  return { csv: inputs.csv };
});
