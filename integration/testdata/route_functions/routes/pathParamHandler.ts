import { RouteFunction } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  return {
    body: JSON.stringify({
      foo: request.params["foo"] || "",
    }),
  };
};

export default handler;
