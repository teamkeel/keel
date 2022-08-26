import { test, expect, Post, logger } from '@teamkeel/testing'

test('create', async () => {
  const p = await Post.create({ title: 'apple' })
  expect.equal(p.title, 'apple')
})

test('update', async () => {
  const post = await Post.create({ title: 'star wars' })

  const updatedPost = await Post.update(post.id, { title: 'star wars sucks!' })

  expect.equal(updatedPost.title, 'star wars sucks!')
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

test('findMany', async () => {
  await Post.create({ title: 'fruit' })
  await Post.create({ title: 'big fruit' })

  const allFruit = await Post.where({
    title: {
      contains: 'fruit'
    }
  }).all()

  expect.equal(allFruit.length, 2)
})

test('chained conditions', async () => {
  await Post.create({ title: 'melon' })
  await Post.create({ title: 'kiwi' })

  const matches = await Post.where({
    title: 'melon'
  }).orWhere({
    title: 'kiwi'
  }).all()

  expect.equal(matches.length, 2)
})

test('order', async () => {
  await Post.create({ title: 'abc' })
  await Post.create({ title: 'bcd' })

  const orderedAlphabetically = await Post.where({
    title: {
      contains: 'bc'
    }
  }).order({
    title: 'desc'
  }).all()

  expect.equal(orderedAlphabetically.length, 2)
  expect.equal(orderedAlphabetically[0].title, 'abc')
})