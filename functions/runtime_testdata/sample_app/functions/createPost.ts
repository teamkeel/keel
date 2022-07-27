import { CreatePostInput, API } from '@teamkeel/sdk'

export default async (inputs: CreatePostInput, api: API) => {
  const { Post } = api.models

  const post = await Post.create(inputs);

  console.log(post.title)

  return {
    post
  }
}
