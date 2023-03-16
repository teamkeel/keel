import { CustomPersonSearch } from "@teamkeel/sdk";

export default CustomPersonSearch(async ({ params }, api, _) => {
  api.permissions.allow();

  const { names } = params;
  const people = await api.models.person.findMany({
    name: { oneOf: names },
  });

  return {
    people,
  };
});
