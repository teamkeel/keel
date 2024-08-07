import { WritePersonUsingCustomFunc, models } from "@teamkeel/sdk";

export default WritePersonUsingCustomFunc(async (ctx, _) => {
  return await models.person.create({ name: ctx.headers.get("Person-Name")! });
});
