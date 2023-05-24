import { models, permissions, CreatePerson } from "@teamkeel/sdk";

export default CreatePerson((ctx, inputs) => {
  permissions.allow();

  let slackId: string | null = null;
  if (inputs.slackId.isNull === undefined || !inputs.slackId.isNull) {
    slackId = inputs.slackId.value!;
  }

  return models.person.create({
    gender: inputs.gender,
    name: inputs.name,
    niNumber: inputs.niNumber,
    slackId: slackId,
  });
});
