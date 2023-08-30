import { CreatePersonWithBeforeWrite, Sex } from "@teamkeel/sdk";

export default CreatePersonWithBeforeWrite({
  beforeWrite: async (ctx, inputs, values) => {
    return {
      ...values,
      title: `${getSalutation(values.sex)} ${values.title}`,
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
