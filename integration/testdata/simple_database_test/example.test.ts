import { test, expect, Post, logger } from '@teamkeel/testing'

test('create', async () => {
  const p = await Post.create({ title: 'apple' })
  expect.equal(p.title, 'apple')
})

test('findOne', async () => {
  await Post.create({ title: 'apple' })
  await Post.create({ title: 'granny apple' })

  const one = await Post.where({
    title: {
      contains: 'apple'
    }
  }).findOne()

  expect.equal(one.title, 'apple')
})

