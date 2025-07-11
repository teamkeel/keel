import { RouteFunction, models } from "@teamkeel/sdk";

const handler: RouteFunction = async (request, ctx) => {
  const person = await models.person.create({
    email: "test@keel.so",
    name: "Test User",
  });

  return {
    body: JSON.stringify(person),
  };
};

export default handler;
