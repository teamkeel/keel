import { test, expect, Actions } from '@teamkeel/testing'

test('destructuring api', async () => {
  const { object: post } = await Actions.createPost({ title: 'apple' })

  expect.equal(post.title, 'apple')
})

