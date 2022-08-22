import { test, expect, Actions } from '@teamkeel/testing'

test('create action', async () => {
  const result = await Actions.createPost({ title: 'foo' })
  expect.equal(result.title, 'foo')
})

test('get action', async () => {
  const post = await Actions.createPost({ title: 'foo' })
  const result = await Actions.getPost({ id: post.id })

  expect.equal(result.id, post.id)
})
