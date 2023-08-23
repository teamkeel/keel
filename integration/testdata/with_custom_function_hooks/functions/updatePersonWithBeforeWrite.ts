import { UpdatePersonWithBeforeWrite, Sex } from "@teamkeel/sdk";

export default UpdatePersonWithBeforeWrite({
  beforeWrite: async (ctx, inputs) => {
    return {
      sex: inputs.values.sex,
      title: `${getSalutation(inputs.values.sex!)} ${inputs.values.title}`,
    };
  },
});

const getSalutation = (sex: Sex) => {
  switch (sex) {
    case Sex.Female:
      return "Ms.";
    case Sex.Male:
      return "Mr.";
  }
};
