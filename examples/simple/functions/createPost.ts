import { CreatePost } from '@teamkeel/sdk'

export default CreatePost(async (inputs, api) => {

  console.log(api)
  const { Post } = api.models;

  const res = await Post.create(inputs)

  console.log(res)

  return res
})