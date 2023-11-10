import { models, permissions, CustomPersonSearch } from "@teamkeel/sdk";

export default CustomPersonSearch(async (_, { params }) => {
  permissions.allow();

  const { names } = params;
  const people = await models.person.findMany({
    where: {
      name: { oneOf: names },
    },
  });

  return {
    people,
  };
});
