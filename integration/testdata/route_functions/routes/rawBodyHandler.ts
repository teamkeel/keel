import { RouteFunction } from "@teamkeel/sdk";
import { createHash } from "node:crypto";

const handler: RouteFunction = async (request, ctx) => {
  const sha1 = createHash("sha1").update(request.body).digest("hex");

  return {
    body: JSON.stringify({
      sha1,
    }),
  };
};

export default handler;
