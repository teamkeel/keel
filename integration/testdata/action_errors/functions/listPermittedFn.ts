import { ListPermittedFn, models } from '@teamkeel/sdk';

export default ListPermittedFn(async (ctx, inputs) => {
	const books = await models.book.findMany(inputs);
	return books;
});
	