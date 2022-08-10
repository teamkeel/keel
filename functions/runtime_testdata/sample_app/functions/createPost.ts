import { CreatePost } from '@teamkeel/sdk'

export default CreatePost(async (inputs, api) => {
  return {
    id: '123',
    title: 'something',
    createdAt: new Date(),
    updatedAt: new Date()
  }
})

