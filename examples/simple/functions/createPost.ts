import { CreatePost } from '@teamkeel/sdk'

export default CreatePost(async (inputs, api) => {
  return await api.models.post.create(inputs)
})