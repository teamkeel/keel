import { MyJob } from "@teamkeel/sdk";

export default MyJob(async (ctx, inputs) => {
    console.log("Hello " + inputs.name);
});