import { test, expect, Actions, Post } from '@teamkeel/testing'

test('isolate 1', async () => {
  await Post.create({ title: 'apple' })

  
})

test('isolate 2', async () => {
  await Post.create({ title: 'apple' })

  const { collection } = await Post.where({
    title: {
      contains: 'apple'
    }
  }).all()

  expect.equal(collection.length, 1)
})
