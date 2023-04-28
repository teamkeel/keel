import { CreatePerson } from "@teamkeel/sdk";

export default CreatePerson(async (inputs, api, ctx) => {
  api.permissions.allow();
  console.log(
    JSON.stringify(
      {
        hello: "1",
        arr: [
          {
            dkk: "1",
          },
          {
            ddkdjdj: "1",
          },
        ],
      },
      null,
      2
    )
  );
  const result = await api.models.person.create(inputs);

  console.log(result);
  return result;
});
