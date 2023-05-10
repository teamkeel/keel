import { BulkPersonUpload } from "@teamkeel/sdk";

export default BulkPersonUpload(async (_, value, api) => {
  api.permissions.allow();

  return {
    people: await Promise.all(
      value.people.map((p) => api.models.person.create(p))
    ),
  };
});
