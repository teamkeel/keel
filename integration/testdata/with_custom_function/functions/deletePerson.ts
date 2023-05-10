import { DeletePerson } from "@teamkeel/sdk";

export default DeletePerson((ctx, inputs, api) => {
  api.permissions.allow();

  return api.models.person.delete({
    id: inputs.id,
  });
});
