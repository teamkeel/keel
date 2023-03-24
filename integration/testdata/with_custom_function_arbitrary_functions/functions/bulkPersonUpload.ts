import { BulkPersonUpload } from "@teamkeel/sdk";

export default BulkPersonUpload(async (value, api, _) => {
  api.permissions.allow();

  return value;
});
