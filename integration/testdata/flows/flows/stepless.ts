import { Stepless, models } from "@teamkeel/sdk";

export default Stepless({}, async () => {
  const thing = await models.thing.create({
    name: "Keelson",
  });
});
