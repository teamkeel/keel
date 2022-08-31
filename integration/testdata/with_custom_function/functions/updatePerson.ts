import { UpdatePerson } from '@teamkeel/sdk'

export default UpdatePerson(async ({ id, ...inputs }, api) => {
  return await api.models.person.update(id, inputs)
})