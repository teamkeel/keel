import { CustomPersonSearch } from "@teamkeel/sdk";

export default CustomPersonSearch(async (_, { params }, api) => {
  api.permissions.allow();

  const { names } = params;
  const people = await api.models.person.findMany({
    name: { oneOf: names },
  });

  return {
    people,
  };
});
