import { UpdatePerson } from '@teamkeel/sdk'

export default UpdatePerson(async ({ where: { id }, values }, api) => {
  return await api.models.person.update(id, values)
})