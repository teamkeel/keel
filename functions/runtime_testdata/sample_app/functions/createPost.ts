import { CreatePost } from '@teamkeel/sdk'

export default CreatePost(async (inputs, api) => {
  const { Post } = api.models

  const post = await Post.create(inputs);

  console.log(post.title)

  return post
})

