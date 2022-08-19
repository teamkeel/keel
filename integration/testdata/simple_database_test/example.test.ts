import { test, expect, Post } from '@teamkeel/testing'

test('it passes', async () => {
  const api = await Post
  const p = await api.create({ title: 'apple' })
  expect.equal(p.title, 'apple')
})

test('it fails', async () => {
  const api = await Post
  const p = await api.create({ title: 'apple' })
  expect.equal(p.title, 'orange')
})

