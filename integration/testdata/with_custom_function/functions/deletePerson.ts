import { DeletePerson } from "@teamkeel/sdk";

export default DeletePerson((inputs, api, ctx) => {
  api.permissions.allow();

  return api.models.person.delete({
    id: inputs.id,
  });
});
