import { RouteFunction } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  const body = JSON.parse(request.body);

  return {
    body: JSON.stringify({
      foo: body.foo,
      fizz: "buzz",
    }),
  };
};

export default handler;
