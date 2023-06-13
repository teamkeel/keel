import { CreatePerson, models } from '@teamkeel/sdk';

export default CreatePerson(async (ctx, inputs) => {
	const person = await models.person.create(inputs);
	return person;
});
	