import { ListDbPermissionFn, models } from '@teamkeel/sdk';

export default ListDbPermissionFn(async (ctx, inputs) => {
	const books = await models.book.findMany(inputs);
	return books;
});
	