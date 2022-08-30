import { test, expect, Actions } from '@teamkeel/testing'

test('create action', async () => {
  const { object: createdPost } = await Actions.createPost({ title: 'foo' })
  expect.equal(createdPost.title, 'foo')
})

test('get action', async () => {
  const { object: post } = await Actions.createPost({ title: 'foo' })
  const { object: fetchedPost } = await Actions.getPost({ id: post.id })

  expect.equal(fetchedPost.id, post.id)
})
