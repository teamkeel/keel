import { UpdatePerson } from "@teamkeel/sdk";

export default UpdatePerson((inputs, api) => {
  return api.models.person.update(inputs.where, inputs.values);
});
