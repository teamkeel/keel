import { RouteFunction } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  const q = new URLSearchParams(request.query);
  return {
    body: `query someParam = ${q.get("someParam")}`,
  };
};

export default handler;
