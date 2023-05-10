import { CustomSearch } from "@teamkeel/sdk";

export default CustomSearch(async (_, value, api) => {
  api.permissions.allow();

  return value;
});
