import { GetPost } from '@teamkeel/sdk'

export default GetPost(async (inputs, api) => {
  // Build something awesome

  return api.models.post.findOne(inputs)
})