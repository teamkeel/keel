import { CustomSearch } from "@teamkeel/sdk";

export default CustomSearch(async (value, api, _) => {
  api.permissions.allow();

  return value;
});
