import { RouteFunction } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  return {
    body: "",
    statusCode: 204,
  };
};

export default handler;
