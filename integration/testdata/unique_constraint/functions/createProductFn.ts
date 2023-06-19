import { CreateProductFn, models } from "@teamkeel/sdk";

export default CreateProductFn(async (ctx, inputs) => {
  const product = await models.product.create(inputs);
  return product;
});
