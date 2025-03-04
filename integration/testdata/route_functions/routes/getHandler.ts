import { RouteFunction } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  const q = new URLSearchParams(request.query);

  return {
    body: JSON.stringify({
      foo: q.get("foo"),
    }),
  };
};

export default handler;
