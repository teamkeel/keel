import { GetPerson } from "@teamkeel/sdk";

export default GetPerson((inputs, api, ctx) => {
  api.permissions.allow();

  return api.models.person.findOne(inputs);
});
