import { test, expect, Actions } from '@teamkeel/testing'

test('create action', async () => {
  const { object: createdPost } = await Actions.createPost({ title: 'foo' })
  expect.equal(createdPost.title, 'foo')
})

test('create action (unrecognised fields)', async () => {
  const { object: createdPost } = await Actions.createPost({ unknown: 'foo' })

  // todo: replace with errors once we populate them
  expect.equal(createdPost, null)
})

test('get action', async () => {
  const { object: post } = await Actions.createPost({ title: 'foo' })
  const { object: fetchedPost } = await Actions.getPost({ id: post.id })

  expect.equal(fetchedPost.id, post.id)
})

// This test verifies that you can't fetch by a field not specified in action inputs
test('get action (non unique)', async () => {
  const { object: post } = await Actions.createPost({ title: 'foo' })
  const { object: fetchedPost } = await Actions.getPost({ title: post.title })

  // todo: until we return errors for 404s, we want to assert
  // that nothing is returned
  expect.equal(fetchedPost, null)
})
