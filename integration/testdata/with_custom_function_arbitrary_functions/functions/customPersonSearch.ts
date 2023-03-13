import { CustomPersonSearch } from "@teamkeel/sdk";

export default CustomPersonSearch(async ({ params }, api, _) => {
  const { names } = params;
  const people = await api.models.person.findMany({
    name: { oneOf: names },
  });

  return {
    people,
  };
});
