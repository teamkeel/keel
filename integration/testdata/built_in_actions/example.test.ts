import { test, expect, Actions } from '@teamkeel/testing'

test('create action', async () => {
  const { object } = await Actions.createPost({ title: 'foo' })
  expect.equal(object.title, 'foo')
})

test('get action', async () => {
  const { object: post } = await Actions.createPost({ title: 'foo' })
  const { object } = await Actions.getPost({ id: post.id })

  expect.equal(object.id, post.id)
})
