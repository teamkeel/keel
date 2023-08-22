import { CreatePersonWithBeforeWrite, Sex } from "@teamkeel/sdk";

export default CreatePersonWithBeforeWrite({
  beforeWrite: async (ctx, inputs) => {
    return {
      ...inputs,
      title: `${getSalutation(inputs.sex)} ${inputs.title}`,
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
