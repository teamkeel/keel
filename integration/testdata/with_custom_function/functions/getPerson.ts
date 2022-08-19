import { GetPerson } from '@teamkeel/sdk'

export default GetPerson(async (inputs, api) => {
  return await api.models.person.find(inputs)
})