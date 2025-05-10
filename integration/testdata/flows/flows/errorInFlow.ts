import { ErrorInFlow } from "@teamkeel/sdk";

export default ErrorInFlow({}, async (ctx) => {
  throw new Error("Error in flow");
});
