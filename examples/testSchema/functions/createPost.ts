import { CreatePost } from '@teamkeel/sdk'

export default CreatePost(async (inputs, api) => {
  return {
    title: 'A post title',
    id: '123',
    updatedAt: new Date(),
    createdAt: new Date()
  }
})