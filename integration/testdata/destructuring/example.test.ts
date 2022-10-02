import { test, expect, actions } from '@teamkeel/testing'

test('destructuring api', async () => {
  const { object: post } = await actions
    .createPost({ title: 'apple' })

  expect.equal(post.title, 'apple')
})

