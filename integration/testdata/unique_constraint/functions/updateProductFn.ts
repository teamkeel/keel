import { UpdateProductFn, models } from "@teamkeel/sdk";

export default UpdateProductFn(async (ctx, inputs) => {
  const product = await models.product.update(inputs.where, inputs.values!);
  return product;
});
