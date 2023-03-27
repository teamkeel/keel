import { BulkPersonUpload } from "@teamkeel/sdk";

export default BulkPersonUpload(async (value, api, _) => {
  api.permissions.allow();

  return {
    people: await Promise.all(
      value.people.map((p) => api.models.person.create(p))
    ),
  };
});
