import { CustomPermission } from "@teamkeel/sdk";

const GUEST_LIST = ["Adam", "Jon", "Dave"];

export default CustomPermission((ctx, inputs, api) => {
  const { name } = inputs;

  if (!GUEST_LIST.includes(name)) {
    // if your name's not on the list, you're not coming in.
    api.permissions.deny();
  } else {
    // you're alright
    api.permissions.allow();
  }

  return api.models.person.create(inputs);
});
