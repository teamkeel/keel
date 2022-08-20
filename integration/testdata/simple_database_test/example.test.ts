import { test, expect, Post } from '@teamkeel/testing'

test('it passes', async () => {
  const p = await Post.create({ title: 'apple' })
  expect.equal(p.title, 'apple')
})

test('it fails', async () => {
  const p = await Post.create({ title: 'apple' })
  expect.equal(p.title, 'orange')
})

