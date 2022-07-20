import { PostInputs, API } from '@teamkeel/sdk'

export default async (inputs: PostInputs, api: API) => {
  const { Post } = api.models

  const createdPost = await Post.create(inputs);

  console.log(createdPost.title)

  return {
    result: createdPost
  }
}