import { PostInputs, API } from '@teamkeel/sdk'

export default async (inputs: PostInputs, api: API) => {
  const { Post } = api.models

  const post = await Post.create(inputs);

  console.log(post.title)

  return {
    post
  }
}