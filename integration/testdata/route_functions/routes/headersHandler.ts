import { RouteFunction } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  const v = request.headers.get(`X-My-Request-Header`);

  return {
    body: "",
    headers: {
      [`X-My-Response-Header`]: `${v}bar`,
    },
  };
};

export default handler;
