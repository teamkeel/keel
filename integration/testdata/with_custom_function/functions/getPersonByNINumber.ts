import { GetPerson } from "@teamkeel/sdk";

export default GetPerson((inputs, api) => {
  return api.models.person.findOne(inputs);
});
