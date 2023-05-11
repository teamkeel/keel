import { models, UpdatePerson } from "@teamkeel/sdk";

export default UpdatePerson((ctx, inputs) => {
  return models.person.update(inputs.where, inputs.values);
});
