import { RouteFunction, models } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  const body = JSON.parse(request.body);

  const person = await models.person.create({
    name: body.name,
  });

  return {
    body: JSON.stringify({
      id: person.id,
    }),
  };
};

export default handler;
