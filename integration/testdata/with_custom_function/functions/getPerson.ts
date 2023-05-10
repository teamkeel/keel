import { GetPerson } from "@teamkeel/sdk";

export default GetPerson((ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.person.findOne(inputs);
});
