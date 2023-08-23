import { permissions, CustomPermission } from "@teamkeel/sdk";

const GUEST_LIST = ["Adam", "Jon", "Dave"];

export default CustomPermission({
  afterWrite: async (ctx, inputs, data) => {
    const { name } = inputs;

    if (!GUEST_LIST.includes(name)) {
      // if your name's not on the list, you're not coming in.
      permissions.deny();
    } else {
      // you're alright
      permissions.allow();
    }
  },
});
