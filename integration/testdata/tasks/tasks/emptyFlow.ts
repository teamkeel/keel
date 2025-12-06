import { EmptyFlow } from "@teamkeel/sdk";

export default EmptyFlow({}, async (ctx, inputs) => {
    await new Promise((resolve) => setTimeout(resolve, 100));
});
