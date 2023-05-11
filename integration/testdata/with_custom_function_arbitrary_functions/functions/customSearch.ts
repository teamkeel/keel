import { models, permissions, CustomSearch } from "@teamkeel/sdk";

export default CustomSearch(async (_, value) => {
  permissions.allow();

  return value;
});
