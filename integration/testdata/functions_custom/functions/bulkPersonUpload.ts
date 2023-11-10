import { models, permissions, BulkPersonUpload } from "@teamkeel/sdk";

export default BulkPersonUpload(async (_, value) => {
  permissions.allow();

  return {
    people: await Promise.all(value.people.map((p) => models.person.create(p))),
  };
});
