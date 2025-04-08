import { RouteFunction } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  return {
    body: JSON.stringify({
      message: "This is a raw HTTP response",
      timestamp: new Date().toISOString(),
    }),
    statusCode: 200,
    headers: {
      "Content-Type": "application/json",
    },
  };
};

export default handler;
