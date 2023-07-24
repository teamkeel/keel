import { ListNotPermittedFn, models } from '@teamkeel/sdk';

export default ListNotPermittedFn(async (ctx, inputs) => {
	const books = await models.book.findMany(inputs);
	return books;
});
	