import { test, expect, Actions, Post } from '@teamkeel/testing'

test('create action', async () => {
  // todo: /Users/adambull/dev/keel/runtime/actions/create.go:46 ERROR: null value in column "title" of relation "post" violates not-null constraint (SQLSTATE 23502)
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

test('list action - equals', async () => {
  await Post.create({ title: 'apple' })
  await Post.create({ title: 'apple' })

  const { collection } = await Actions.listPosts({ title: { equals: 'apple' } })

  expect.equal(collection.length, 2)
})

test('list action - contains', async () => {
  await Post.create({ title: 'banan' })
  await Post.create({ title: 'banana' })

  const { collection } = await Actions.listPosts({ title: { contains: 'ana' } })

  expect.equal(collection.length, 2)
})

test('list action - startsWith', async () => {
  await Post.create({ title: 'adam' })
  await Post.create({ title: 'adamant' })

  const { collection } = await Actions.listPosts({ title: { startsWith: 'adam' } })

  expect.equal(collection.length, 2)
})

test('list action - endsWith', async () => {
  await Post.create({ title: 'star wars' })
  await Post.create({ title: 'a post about star wars' })

  const { collection } = await Actions.listPosts({ title: { endsWith: 'star wars' } })

  expect.equal(collection.length, 2)
})

test('list action - oneOf', async () => {
  await Post.create({ title: 'pear' })
  await Post.create({ title: 'mango' })

  const { collection } = await Actions.listPosts({ title: { oneOf: ['pear', 'mango'] } })

  expect.equal(collection.length, 2)
})
