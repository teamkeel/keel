import { DeletePerson } from "@teamkeel/sdk";

export default DeletePerson((inputs, api) => {
  return api.models.person.delete({
    id: inputs.id,
  });
});
