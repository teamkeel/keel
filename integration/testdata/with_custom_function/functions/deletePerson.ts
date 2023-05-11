import { models, DeletePerson } from "@teamkeel/sdk";

export default DeletePerson((ctx, inputs) => {
  return models.person.delete({
    id: inputs.id,
  });
});
