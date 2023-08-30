import { UpdatePersonWithBeforeWrite, Sex } from "@teamkeel/sdk";

export default UpdatePersonWithBeforeWrite({
  beforeWrite: async (ctx, inputs, values) => {
    return {
      sex: values.sex,
      title: `${getSalutation(values.sex!)} ${values.title}`,
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
