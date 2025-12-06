import { EmptyFlow } from "@teamkeel/sdk";

export default EmptyFlow({}, async (ctx) => {
  await new Promise((resolve) => setTimeout(resolve, 100));
});
