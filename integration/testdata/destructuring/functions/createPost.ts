import { CreatePost } from '@teamkeel/sdk'

export default CreatePost(async (inputs, api) => {
  const { post } = api.models
  return await post.create(inputs)
})