import { models, GetPerson } from "@teamkeel/sdk";

export default GetPerson((ctx, inputs) => {
  return models.person.findOne(inputs);
});
