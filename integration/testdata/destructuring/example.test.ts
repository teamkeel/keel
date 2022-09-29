import { test, expect, actions } from '@teamkeel/testing'

test('destructuring api', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPost({ title: 'apple' })

  expect.equal(post.title, 'apple')
})

