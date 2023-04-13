import { UpdatePost } from '@teamkeel/sdk';

export default UpdatePost(async (inputs, api, ctx) => {
	const post = await api.models.post.update(inputs.where, inputs.values);
	return post;
});
	